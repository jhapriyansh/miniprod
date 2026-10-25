package pool

import "testing"

func TestPick(t *testing.T) {
    tests := []struct {
        name    string
        entries map[string]Entry
        skip    map[string]bool
        want    string
        wantErr bool
    }{
        {
            name: "healthy only, least active wins",
            entries: map[string]Entry{
                "w1": {State: Healthy, Active: 3},
                "w2": {State: Healthy, Active: 1},
            },
            want: "w2",
        },
        {
            name: "tie breaks by lowest ID",
            entries: map[string]Entry{
                "w2": {State: Healthy, Active: 1},
                "w1": {State: Healthy, Active: 1},
            },
            want: "w1",
        },
        {
            name: "healthy beats suspect even with fewer connections",
            entries: map[string]Entry{
                "w1": {State: Suspect, Active: 0},
                "w2": {State: Healthy, Active: 5},
            },
            want: "w2",
        },
        {
            name: "healthy beats dead even with fewer connections",
            entries: map[string]Entry{
                "w1": {State: Dead, Active: 0},
                "w2": {State: Healthy, Active: 9},
            },
            want: "w2",
        },
        {
            name: "no healthy, only suspect and dead",
            entries: map[string]Entry{
                "w1": {State: Suspect},
                "w2": {State: Dead},
            },
            wantErr: true,
        },
        {
            name:    "no backends at all",
            entries: map[string]Entry{},
            wantErr: true,
        },
        {
            name: "skipped backend is never picked",
            entries: map[string]Entry{
                "w1": {State: Healthy, Active: 0},
                "w2": {State: Healthy, Active: 5},
            },
            skip: map[string]bool{"w1": true},
            want: "w2",
        },
        {
            name: "all healthy backends skipped returns no healthy",
            entries: map[string]Entry{
                "w1": {State: Healthy, Active: 0},
            },
            skip:    map[string]bool{"w1": true},
            wantErr: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Pick(tc.entries, tc.skip)
            if tc.wantErr {
                if err != ErrNoHealthy {
                    t.Fatalf("got (%q, %v), want ErrNoHealthy", got, err)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tc.want {
                t.Errorf("Pick = %q, want %q", got, tc.want)
            }
        })
    }
}
