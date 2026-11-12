package main

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/jhapriyansh/miniprod/internal/pool"
	"github.com/jhapriyansh/miniprod/internal/proxy"
)

// 503 sent to the client when no backend can be reached. The load generator
// speaks HTTP, so this one line is HTTP-aware.
const unavailable = "HTTP/1.1 503 Service Unavailable\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"

type balancer struct {
	mu      sync.Mutex
	entries map[string]pool.Entry // keyed by backend address (host:port)
}

// pick is the proxy.PickFunc. It holds the lock while reading the pool.
func (b *balancer) pick(skip map[string]bool) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return pool.Pick(b.entries, skip)
}

// adjust changes a backend's active connection count under the lock.
func (b *balancer) adjust(addr string, delta int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e := b.entries[addr]
	e.Active += delta
	b.entries[addr] = e
}

func (b *balancer) handle(client net.Conn) {
	be, addr, err := proxy.Connect(b.pick, proxy.DefaultDialer, proxy.MaxDialAttempts)
	if err != nil {
		io.WriteString(client, unavailable)
		client.Close()
		return
	}
	b.adjust(addr, +1)
	defer b.adjust(addr, -1)
	proxy.Relay(client, be)
}

func main() {
	b := &balancer{entries: make(map[string]pool.Entry)}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("balancer listening on %s", ln.Addr())
	log.Fatal(proxy.Serve(ln, b.handle))
}
