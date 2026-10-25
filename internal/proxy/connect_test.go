package proxy

import (
    "errors"
    "net"
    "testing"
)

// fakeDialer records every dial and refuses the addresses in refuse.
// It returns an in-memory connection instead of opening a socket.
type fakeDialer struct {
    refuse map[string]bool
    dials  []string
}

func (f *fakeDialer) Dial(addr string) (net.Conn, error) {
    f.dials = append(f.dials, addr)
    if f.refuse[addr] {
        return nil, errors.New("connection refused")
    }
    client, server := net.Pipe()
    server.Close()
    return client, nil
}

// fakePick returns the first address not in skip, in the order given.
func fakePick(addrs ...string) PickFunc {
    return func(skip map[string]bool) (string, error) {
        for _, a := range addrs {
            if !skip[a] {
                return a, nil
            }
        }
        return "", errors.New("no healthy backends")
    }
}

func TestConnect(t *testing.T) {
    tests := []struct {
        name      string
        backends  []string
        refuse    []string
        wantDials int
        wantAddr  string
        wantErr   error
        anyErr    bool
    }{
        {
            name:      "no refusals, one dial",
            backends:  []string{"a", "b"},
            wantDials: 1,
            wantAddr:  "a",
        },
        {
            name:      "one refusal, retry succeeds on a different backend",
            backends:  []string{"a", "b"},
            refuse:    []string{"a"},
            wantDials: 2,
            wantAddr:  "b",
        },
        {
            name:      "two refusals, gives up with unavailable",
            backends:  []string{"a", "b", "c"},
            refuse:    []string{"a", "b"},
            wantDials: 2,
            wantErr:   ErrUnavailable,
        },
        {
            name:      "no backends at all, no dial made",
            backends:  []string{},
            wantDials: 0,
            anyErr:    true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            refuse := make(map[string]bool)
            for _, a := range tc.refuse {
                refuse[a] = true
            }
            d := &fakeDialer{refuse: refuse}

            conn, addr, err := Connect(fakePick(tc.backends...), d.Dial, maxDialAttempts)
            if conn != nil {
                defer conn.Close()
            }

            if got := len(d.dials); got != tc.wantDials {
                t.Fatalf("dials = %d, want %d (dialed %v)", got, tc.wantDials, d.dials)
            }

            switch {
            case tc.anyErr:
                if err == nil {
                    t.Fatal("want an error, got nil")
                }
            case tc.wantErr != nil:
                if !errors.Is(err, tc.wantErr) {
                    t.Fatalf("err = %v, want %v", err, tc.wantErr)
                }
            default:
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
                if addr != tc.wantAddr {
                    t.Errorf("addr = %q, want %q", addr, tc.wantAddr)
                }
            }
        })
    }
}
