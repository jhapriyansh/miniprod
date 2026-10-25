package pool

import (
	"errors"
	"sort"
)

var ErrNoHealthy = errors.New("no healthy backends")

func Pick(entries map[string]Entry) (string, error) {
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var best string
	var bestActive int
	found := false

	for _, id := range ids {
		e := entries[id]
		if e.State != Healthy {
			continue
		}
		if !found || e.Active < bestActive {
			best, bestActive, found = id, e.Active, true
		}
	}

	if !found {
		return "", ErrNoHealthy
	}

	return best, nil
}
