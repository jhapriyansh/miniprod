package pool

import "time"

type State int

const (
	Healthy State = iota
	Suspect
	Dead
)

func (s State) String() string {
	switch s {
	case Healthy:
		return "Healthy"
	case Suspect:
		return "Suspect"
	case Dead:
		return "Dead"
	default:
		return "Unknown"
	}
}

type Params struct {
	T time.Duration // beat interval
	S int           // missed beats before SUSPECT
	N int           // missed beats before DEAD
	R int           // consecutive beats to recover one level
}

type Entry struct {
	State    State
	LastSeen time.Time // set on every valid beat
	Streak   int       // consecutive valid beats while not HEALTHY
	Active   int       // proxy conns open to this backend
}

// OnBeat handles a valid heartbeat arriving at now.
func OnBeat(e Entry, now time.Time, p Params) Entry {
	silence := now.Sub(e.LastSeen)
	e.LastSeen = now

	if e.State == Healthy {
		e.Streak = 0
		return e
	}

	// A beat more than 2×T after the previous one breaks the run,
	// so this beat starts a new streak.
	if silence > time.Duration(2)*p.T {
		e.Streak = 1
	} else {
		e.Streak++
	}

	if e.Streak >= p.R {
		switch e.State {
		case Suspect:
			e.State = Healthy
		case Dead:
			e.State = Suspect
		}
		e.Streak = 0
	}
	return e
}

// OnTick runs periodically with no packet, and checks silence since LastSeen.
// Downward transitions move one level per tick: HEALTHY -> SUSPECT -> DEAD.
func OnTick(e Entry, now time.Time, p Params) Entry {
	silence := now.Sub(e.LastSeen)

	switch {
	case silence >= time.Duration(p.N)*p.T && e.State == Suspect:
		e.State = Dead
		e.Streak = 0
	case silence >= time.Duration(p.S)*p.T && e.State == Healthy:
		e.State = Suspect
		e.Streak = 0
	case silence > time.Duration(2)*p.T && e.State != Healthy:
		e.Streak = 0
	}
	return e
}
