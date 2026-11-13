package pool

import (
	"sync"
	"time"
)

// Store is the balancer's shared backend table. Every access holds mu.
type Store struct {
	mu      sync.Mutex
	params  Params
	entries map[string]Entry
	last    string // key returned by the previous Pick, for tie rotation
}

func NewStore(p Params) *Store {
	return &Store{params: p, entries: make(map[string]Entry)}
}

// Beat records a valid heartbeat from addr. A backend seen for the first
// time starts as SUSPECT with no LastSeen, so it must earn HEALTHY.
func (s *Store) Beat(addr string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		e = Entry{State: Suspect}
	}
	s.entries[addr] = OnBeat(e, now, s.params)
}

// Tick runs the silence check for every known backend.
func (s *Store) Tick(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, e := range s.entries {
		s.entries[addr] = OnTick(e, now, s.params)
	}
}

// Pick is the proxy's PickFunc: least active connections among healthy backends.
// Pick is the proxy's PickFunc. It records the choice so the next tie rotates.
func (s *Store) Pick(skip map[string]bool) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, err := Pick(s.entries, skip, s.last)
	if err == nil {
		s.last = id
	}
	return id, err
}

// Adjust changes a backend's active connection count. Unknown addresses are
// ignored: a zero-value Entry is HEALTHY, so creating one here would route traffic.
func (s *Store) Adjust(addr string, delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		return
	}
	e.Active += delta
	s.entries[addr] = e
}
