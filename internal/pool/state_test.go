package pool

import (
	"testing"
	"time"
)

var base = time.Unix(0, 0)
var p = Params{T: time.Second, S: 2, N: 5, R: 3}

type step struct {
	op         string // "beat" or "tick"
	at         time.Duration
	wantState  State
	wantStreak int
}

func beat(at time.Duration, s State, streak int) step {
	return step{"beat", at, s, streak}
}

func tick(at time.Duration, s State, streak int) step {
	return step{"tick", at, s, streak}
}

func TestStateMachine(t *testing.T) {
	cases := []struct {
		name  string
		start Entry
		steps []step
	}{
		{
			name:  "new entry, first beat",
			start: Entry{State: Suspect},
			steps: []step{beat(1*time.Second, Suspect, 1)},
		},
		{
			name:  "recovery completes",
			start: Entry{State: Suspect, LastSeen: base, Streak: 2},
			steps: []step{beat(1*time.Second, Healthy, 0)},
		},
		{
			name:  "beat while healthy changes nothing",
			start: Entry{State: Healthy, LastSeen: base},
			steps: []step{beat(1*time.Second, Healthy, 0)},
		},
		{
			name:  "silence moves healthy to suspect",
			start: Entry{State: Healthy, LastSeen: base},
			steps: []step{tick(2*time.Second, Suspect, 0)},
		},
		{
			name:  "silence moves suspect to dead",
			start: Entry{State: Suspect, LastSeen: base},
			steps: []step{tick(5*time.Second, Dead, 0)},
		},
		{
			name:  "healthy silent past N goes to suspect first",
			start: Entry{State: Healthy, LastSeen: base},
			steps: []step{
				tick(5*time.Second, Suspect, 0),
				tick(5*time.Second, Dead, 0),
			},
		},
		{
			name:  "no direct dead to healthy",
			start: Entry{State: Dead, LastSeen: base, Streak: 2},
			steps: []step{beat(1*time.Second, Suspect, 0)},
		},
		{
			name:  "full dead recovery passes through suspect",
			start: Entry{State: Dead, LastSeen: base},
			steps: []step{
				beat(1*time.Second, Dead, 1),
				beat(2*time.Second, Dead, 2),
				beat(3*time.Second, Suspect, 0),
				beat(4*time.Second, Suspect, 1),
				beat(5*time.Second, Suspect, 2),
				beat(6*time.Second, Healthy, 0),
			},
		},
		{
			name:  "silence breaks a recovery streak",
			start: Entry{State: Suspect, LastSeen: base, Streak: 2},
			steps: []step{tick(3*time.Second, Suspect, 0)},
		},
		{
			name:  "late beat within jitter tolerance keeps streak",
			start: Entry{State: Suspect, LastSeen: base, Streak: 1},
			steps: []step{beat(1500*time.Millisecond, Suspect, 2)},
		},
		{
			// Silence is over 2×T, so this beat starts a new streak at 1.
			name:  "beat 2.5s late restarts streak at 1",
			start: Entry{State: Suspect, LastSeen: base, Streak: 1},
			steps: []step{beat(2500*time.Millisecond, Suspect, 1)},
		},
		{
			name:  "dead needs full silence, not 4.9s",
			start: Entry{State: Healthy, LastSeen: base},
			steps: []step{tick(4900*time.Millisecond, Suspect, 0)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.start
			for i, s := range tc.steps {
				now := base.Add(s.at)
				switch s.op {
				case "beat":
					e = OnBeat(e, now, p)
				case "tick":
					e = OnTick(e, now, p)
				default:
					t.Fatalf("step %d: unknown op %q", i, s.op)
				}
				if e.State != s.wantState || e.Streak != s.wantStreak {
					t.Fatalf("step %d (%s at %v): got %v streak %d, want %v streak %d",
						i, s.op, s.at, e.State, e.Streak, s.wantState, s.wantStreak)
				}
			}
		})
	}
}
