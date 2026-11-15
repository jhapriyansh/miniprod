package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// parseStub reads nginx stub_status output and returns active connections
// and the total number of requests handled.
func parseStub(body string) (int, uint64, error) {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	active := -1
	var requests uint64
	haveReq := false
	for i, l := range lines {
		if strings.HasPrefix(l, "Active connections:") {
			f := strings.Fields(l)
			if len(f) < 3 {
				return 0, 0, fmt.Errorf("short active line %q", l)
			}
			n, err := strconv.Atoi(f[2])
			if err != nil {
				return 0, 0, err
			}
			active = n
		}
		if strings.HasPrefix(l, "server accepts") && i+1 < len(lines) {
			f := strings.Fields(lines[i+1])
			if len(f) < 3 {
				return 0, 0, fmt.Errorf("short counters line %q", lines[i+1])
			}
			r, err := strconv.ParseUint(f[2], 10, 64)
			if err != nil {
				return 0, 0, err
			}
			requests = r
			haveReq = true
		}
	}
	if active < 0 || !haveReq {
		return 0, 0, fmt.Errorf("stub_status output incomplete")
	}
	return active, requests, nil
}

func poll(c *http.Client, url string) (int, uint64, error) {
	resp, err := c.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	return parseStub(string(body))
}

// appHealthy is the v2 gate: true only when the app answers 200 within the
// client's timeout. A frozen worker fails this, so the sidecar stops beating.
func appHealthy(client *http.Client, url string) bool {
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	pod := os.Getenv("POD_NAME")
	podIP := os.Getenv("POD_IP")
	balancer := os.Getenv("BALANCER_ADDR")
	if pod == "" || podIP == "" || balancer == "" {
		log.Fatal("POD_NAME, POD_IP and BALANCER_ADDR are required")
	}
	statusURL := getenv("STATUS_URL", "http://127.0.0.1:8081/stub_status")
	appURL := getenv("APP_URL", "http://127.0.0.1:8080/")
	addr := net.JoinHostPort(podIP, "8080")

	conn, err := net.Dial("udp", balancer)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{Timeout: 500 * time.Millisecond}
	probeClient := &http.Client{Timeout: 1 * time.Second}
	ticker := time.NewTicker(time.Second)

	var seq, lastReq uint64
	var lastAt time.Time
	for now := range ticker.C {
		seq++

		active, reqs, err := poll(client, statusURL)
		rps := 0.0
		if err != nil {
			log.Printf("stub_status: %v", err)
			active = 0
		} else if !lastAt.IsZero() && reqs >= lastReq {
			rps = float64(reqs-lastReq) / now.Sub(lastAt).Seconds()
			lastReq = reqs
		} else {
			lastReq = reqs
		}
		lastAt = now

		// v2: beat only when the app itself answers. A frozen worker
		// stops the beats, so the balancer's silence detection takes over.
		if !appHealthy(probeClient, appURL) {
			log.Printf("app probe failed, no beat sent")
			continue
		}
		fmt.Fprintf(conn, "v=1 id=%s addr=%s seq=%d conns=%d rps=%.2f",
			pod, addr, seq, active, rps)
	}
}
