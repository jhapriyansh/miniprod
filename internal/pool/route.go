package pool

import (
	"errors"
	"sort"
)

// ErrNoHealthy means no backend can take a new connection.
// The proxy turns this into an HTTP 503.
var ErrNoHealthy = errors.New("no healthy backends")

// Pick returns the Healthy backend with the fewest active connections.
// Among equal counts, it rotates: it starts after the key returned last
// time, so sequential requests spread across equal backends instead of
// always landing on the lowest key. Backends in skip are never returned.
// last is the key returned by the previous call, or "" for the first call.
func Pick(entries map[string]Entry, skip map[string]bool, last string) (string, error) {
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Build the candidate list in sorted order, with the state and skip checks applied.
	var cands []string
	for _, id := range ids {
		if skip[id] || entries[id].State != Healthy {
			continue
		}
		cands = append(cands, id)
	}
	if len(cands) == 0 {
		return "", ErrNoHealthy
	}

	// Start scanning just after last, so ties rotate from there.
	start := 0
	for i, id := range cands {
		if id > last {
			start = i
			break
		}
	}

	best := cands[start]
	bestActive := entries[best].Active
	for k := 1; k < len(cands); k++ {
		id := cands[(start+k)%len(cands)]
		if a := entries[id].Active; a < bestActive {
			best, bestActive = id, a
		}
	}
	return best, nil
}
