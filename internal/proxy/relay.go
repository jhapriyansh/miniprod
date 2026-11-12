package proxy

import (
	"io"
	"net"
	"sync"
)

// Relay copies bytes both ways between a and b until both directions finish.
// When one direction reaches EOF, it half-closes the other side, so the peer
// sees EOF but can still send its response. If a copy fails, both connections
// are closed, which unblocks the other direction. Both are closed at the end.
func Relay(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go copyOne(&wg, a, b) // a -> b
	go copyOne(&wg, b, a) // b -> a
	wg.Wait()
	a.Close()
	b.Close()
}

func copyOne(wg *sync.WaitGroup, src, dst net.Conn) {
	defer wg.Done()
	_, err := io.Copy(dst, src)
	if err != nil {
		// A reset or broken pipe: tear down both sides now.
		src.Close()
		dst.Close()
		return
	}
	// Clean EOF: pass the end-of-stream on, but keep the reverse direction open.
	if tc, ok := dst.(*net.TCPConn); ok {
		tc.CloseWrite()
	} else {
		dst.Close()
	}
}
