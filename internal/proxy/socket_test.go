package proxy

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestServeReturnsNilOnClose(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- Serve(ln, func(c net.Conn) { c.Close() }) }()

	ln.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after the listener closed")
	}
}

func TestRelayEchoThroughBalancer(t *testing.T) {
	// Backend: echoes everything back, then closes.
	backend, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	go Serve(backend, func(c net.Conn) {
		defer c.Close()
		io.Copy(c, c)
	})

	// Balancer: accepts a client, dials the backend, relays.
	lb, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lb.Close()
	beAddr := backend.Addr().String()
	go Serve(lb, func(client net.Conn) {
		be, err := DefaultDialer(beAddr)
		if err != nil {
			client.Close()
			return
		}
		Relay(client, be)
	})

	c, err := net.Dial("tcp", lb.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(5 * time.Second))

	if _, err := c.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	// Half-close: tell the balancer we're done sending, but keep reading.
	c.(*net.TCPConn).CloseWrite()

	got, err := io.ReadAll(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestDefaultDialerRefusesClosedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // nothing listens here now

	if conn, err := DefaultDialer(addr); err == nil {
		conn.Close()
		t.Fatal("dial to a closed port succeeded")
	}
}
