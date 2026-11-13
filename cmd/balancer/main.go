package main

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/jhapriyansh/miniprod/internal/pool"
	"github.com/jhapriyansh/miniprod/internal/proxy"
	"github.com/jhapriyansh/miniprod/internal/telemetry"
)

const (
	proxyListen     = ":8080"
	telemetryListen = ":9000"
)

var params = pool.Params{T: time.Second, S: 2, N: 5, R: 3}

// unavailable is the 503 sent when no backend can be reached.
const unavailable = "HTTP/1.1 503 Service Unavailable\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"

type balancer struct {
	store *pool.Store
}

func (b *balancer) handle(client net.Conn) {
	be, addr, err := proxy.Connect(b.store.Pick, proxy.DefaultDialer, proxy.MaxDialAttempts)
	if err != nil {
		io.WriteString(client, unavailable)
		client.Close()
		return
	}
	b.store.Adjust(addr, +1)
	defer b.store.Adjust(addr, -1)
	proxy.Relay(client, be)
}

func main() {
	store := pool.NewStore(params)

	tln, err := net.ListenPacket("udp", telemetryListen)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := telemetry.Receive(tln, store); err != nil {
			log.Fatal(err)
		}
	}()

	// The silence check runs once per beat interval.
	go func() {
		t := time.NewTicker(params.T)
		for now := range t.C {
			store.Tick(now)
		}
	}()

	ln, err := net.Listen("tcp", proxyListen)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("balancer: proxy on %s, telemetry on %s", proxyListen, telemetryListen)
	b := &balancer{store: store}
	log.Fatal(proxy.Serve(ln, b.handle))
}
