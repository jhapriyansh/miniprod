package proxy

import (
    "errors"
    "fmt"
    "net"
)

// ErrUnavailable means every attempt to reach a backend was refused.
// The proxy turns it into an HTTP 503 before any byte reaches the client.
var ErrUnavailable = errors.New("no backend reachable")

// maxDialAttempts is the first dial plus one retry.
const maxDialAttempts = 2

// PickFunc chooses a backend address, skipping any addresses in skip.
type PickFunc func(skip map[string]bool) (string, error)

// Dialer opens a connection to a backend address.
type Dialer func(addr string) (net.Conn, error)

// Connect tries to reach a backend. A refused dial is retried with a
// different backend, up to maxAttempts dials in total. Only the dial is
// retried: once a connection exists, the proxy never tries again.
func Connect(pick PickFunc, dial Dialer, maxAttempts int) (net.Conn, string, error) {
    skip := make(map[string]bool)
    var lastErr error

    for attempt := 0; attempt < maxAttempts; attempt++ {
        addr, err := pick(skip)
        if err != nil {
            // No backend left to try. The caller maps this to a 503 too.
            return nil, "", err
        }

        conn, err := dial(addr)
        if err == nil {
            return conn, addr, nil
        }

        // Don't pick the same refused backend again on the retry.
        skip[addr] = true
        lastErr = err
    }

    return nil, "", fmt.Errorf("%w: last error: %v", ErrUnavailable, lastErr)
}
