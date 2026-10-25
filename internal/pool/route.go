package pool

import (
    "errors"
    "sort"
)

// ErrNoHealthy means no backend can take a new connection.
// The proxy turns this into an HTTP 503.
var ErrNoHealthy = errors.New("no healthy backends")

// Pick returns the ID of the Healthy backend with the fewest active
// connections. Ties go to the lowest ID. Backends listed in skip are
// never returned. It returns ErrNoHealthy if no backend qualifies.
func Pick(entries map[string]Entry, skip map[string]bool) (string, error) {
    // Map iteration order is random. Sorting the IDs makes ties deterministic.
    ids := make([]string, 0, len(entries))
    for id := range entries {
        ids = append(ids, id)
    }
    sort.Strings(ids)

    var best string
    var bestActive int
    found := false

    for _, id := range ids {
        if skip[id] {
            continue
        }
        e := entries[id]
        if e.State != Healthy {
            continue
        }
        // Strict "<" keeps the first (lowest) ID on a tie.
        if !found || e.Active < bestActive {
            best, bestActive, found = id, e.Active, true
        }
    }

    if !found {
        return "", ErrNoHealthy
    }
    return best, nil
}
