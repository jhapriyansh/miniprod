package telemetry

import (
    "errors"
    "log"
    "net"
    "time"
)

// Beater receives a valid heartbeat. The pool store implements it.
type Beater interface {
    Beat(addr string, now time.Time)
}

// Receive reads telemetry datagrams until conn is closed. Invalid packets
// are logged and dropped, never forwarded.
func Receive(conn net.PacketConn, sink Beater) error {
    buf := make([]byte, maxLineLen+1) // one extra byte so Parse sees overlong lines
    for {
        n, _, err := conn.ReadFrom(buf)
        if err != nil {
            if errors.Is(err, net.ErrClosed) {
                return nil
            }
            log.Printf("telemetry read: %v", err)
            continue
        }
        r, err := Parse(string(buf[:n]))
        if err != nil {
            log.Printf("drop packet: %v", err)
            continue
        }
        sink.Beat(r.Addr, time.Now())
    }
}