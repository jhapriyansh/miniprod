package pool

import (
	"testing"
	"time"
)

func TestStoreRoutesAfterRecovery(t *testing.T) {
	s := NewStore(Params{T: time.Second, S: 2, N: 5, R: 3})
	base := time.Unix(0, 0)
	addr := "10.42.1.5:8080"

	s.Beat(addr, base.Add(1*time.Second))
	s.Beat(addr, base.Add(2*time.Second))
	if _, err := s.Pick(nil); err != ErrNoHealthy {
		t.Fatalf("after 2 beats: want ErrNoHealthy, got %v", err)
	}

	s.Beat(addr, base.Add(3*time.Second))
	got, err := s.Pick(nil)
	if err != nil || got != addr {
		t.Fatalf("after 3 beats: got (%q, %v), want %q", got, err, addr)
	}
}

func TestAdjustUnknownAddrDoesNotCreateEntry(t *testing.T) {
	s := NewStore(Params{T: time.Second, S: 2, N: 5, R: 3})
	s.Adjust("ghost:8080", +1)
	if _, err := s.Pick(nil); err != ErrNoHealthy {
		t.Fatalf("want ErrNoHealthy, got %v", err)
	}
}
