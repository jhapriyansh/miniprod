package proxy

import (
	"net"
	"time"
)

// DialTimeout bounds how long a connect to a backend may take.
const DialTimeout = 2 * time.Second

// DefaultDialer is the real Dialer. It fails fast on an unreachable backend
// so the retry in Connect can try another one.
func DefaultDialer(addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, DialTimeout)
}
