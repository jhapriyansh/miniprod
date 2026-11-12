package proxy

import (
	"errors"
	"log"
	"net"
	"time"
)

// Serve accepts connections until the listener is closed, handing each
// one to handle in its own goroutine. It returns nil on a normal close.
func Serve(ln net.Listener, handle func(net.Conn)) error {
	var delay time.Duration
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			// Temporary failure (e.g. out of file descriptors): back off and keep serving.
			if delay == 0 {
				delay = 5 * time.Millisecond
			} else {
				delay *= 2
			}
			if delay > time.Second {
				delay = time.Second
			}
			log.Printf("accept: %v (retrying in %v)", err, delay)
			time.Sleep(delay)
			continue
		}
		delay = 0
		go handle(conn)
	}
}
