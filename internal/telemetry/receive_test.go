package telemetry

import (
    "net"
    "testing"
    "time"
)

type recorder struct{ got chan string }

func (r recorder) Beat(addr string, now time.Time) { r.got <- addr }

func TestReceiveForwardsValidBeatsOnly(t *testing.T) {
    pc, err := net.ListenPacket("udp", "127.0.0.1:0")
    if err != nil {
        t.Fatal(err)
    }
    rec := recorder{got: make(chan string, 4)}
    done := make(chan error, 1)
    go func() { done <- Receive(pc, rec) }()

    c, err := net.Dial("udp", pc.LocalAddr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer c.Close()

    c.Write([]byte("v=1 id=p1 addr=10.42.1.5:8080 seq=1 conns=0 rps=1.0"))
    c.Write([]byte("garbage"))

    select {
    case a := <-rec.got:
        if a != "10.42.1.5:8080" {
            t.Errorf("addr = %q, want 10.42.1.5:8080", a)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("valid beat was not forwarded")
    }

    select {
    case a := <-rec.got:
        t.Fatalf("invalid packet was forwarded as %q", a)
    case <-time.After(200 * time.Millisecond):
    }

    pc.Close()
    if err := <-done; err != nil {
        t.Fatalf("Receive returned %v, want nil", err)
    }
}