package owtracker

import "time"

// rankWatch turns per-frame progress-screen reads into at most one rank per match.
//
//   - it starts when a win/loss is counted and looks for the screen for
//     rankSearchFor (the screen shows after Play of the Game, ~50 s later, or at
//     once if the player leaves); once the screen is up it is followed until it
//     closes, even past that, up to rankWatchMax
//   - on a rank change the screen first shows the old rank while the bar fills,
//     so the latest read seen rankConfirmReads times in a row wins, not the first
//   - it commits when the screen closes (no role for rankGoneAfter), or once the
//     latest read has held for rankSettle with the screen still up; a single read
//     counts if nothing better was seen (players who click Continue fast leave
//     it up ~2 s)
//   - when the rank before the match is known, a win can only keep it or move it
//     up one division and a loss keep it or move it down one: with no read
//     confirmed, a lone read in that set beats other lone reads. A confirmed
//     read still wins: the stored rank is stale after a match whose screen the
//     player skipped
const (
	rankSearchFor    = flowMinMatchGap
	rankWatchMax     = 3 * time.Minute
	rankConfirmReads = 2
	rankSettle       = 15 * time.Second
	rankGoneAfter    = 2 * time.Second
	rankPoll         = 250 * time.Millisecond // one frame per rankPoll, work included
	rankMinPause     = 40 * time.Millisecond  // breather when a read takes longer
	rankBurstFor     = 2 * time.Second        // back to back frames once the screen shows
)

type rankWatch struct {
	started time.Time
	seenAt  time.Time // last frame with the role read (screen up)
	firstAt time.Time // first frame with the role read
	votes   map[RankRead]int
	order   []RankRead // first-seen order, for ties
	last    RankRead   // latest valid read and how many times in a row
	lastN   int
	lastAt  time.Time
	cand    RankRead // latest read confirmed rankConfirmReads times in a row

	expect func(RankRead) bool // reads possible after this match (nil = unknown)
	exp    RankRead            // latest expected read
}

// start opens a watch; expect (may be nil) tells the ranks this match can end on.
func (w *rankWatch) start(now time.Time, expect func(RankRead) bool) {
	*w = rankWatch{started: now, votes: map[RankRead]int{}, expect: expect}
}

// RankExpect is the set of ranks a match can end on: the rank before it (prior,
// a widget ladder index) or one division up after a win, down after a loss.
func RankExpect(outcome Outcome, prior func(role string) (int, bool)) func(RankRead) bool {
	if prior == nil {
		return nil
	}
	step := 1
	if outcome == OutcomeLoss {
		step = -1
	}
	return func(r RankRead) bool {
		before, ok := prior(r.Role)
		return ok && (r.Rank == before || r.Rank == before+step)
	}
}

// active stays true until observe has run past the end (so a lone read is
// still committed).
func (w *rankWatch) active() bool { return w.votes != nil }

func (w *rankWatch) stop() { *w = rankWatch{} }

// screenUp tells whether the progress screen was read within rankGoneAfter.
func (w *rankWatch) screenUp(now time.Time) bool {
	return !w.seenAt.IsZero() && now.Sub(w.seenAt) < rankGoneAfter
}

// burst tells whether the screen showed less than rankBurstFor ago.
func (w *rankWatch) burst(now time.Time) bool {
	return !w.firstAt.IsZero() && now.Sub(w.firstAt) < rankBurstFor
}

// observe returns the confirmed rank exactly once per watch.
func (w *rankWatch) observe(now time.Time, r RankRead) (RankRead, bool) {
	if w.votes == nil {
		return RankRead{}, false
	}
	if r.Role != "" {
		w.seenAt = now
		if w.firstAt.IsZero() {
			w.firstAt = now
		}
	}
	if r.Valid() {
		key := RankRead{Role: r.Role, Tier: r.Tier, Division: r.Division, Rank: r.Rank}
		if w.votes[key] == 0 {
			w.order = append(w.order, key)
		}
		w.votes[key]++
		if key != w.last {
			w.last, w.lastN, w.lastAt = key, 0, now
		}
		w.lastN++
		if w.expect != nil && w.expect(key) {
			w.exp = key
		}
		if w.lastN >= rankConfirmReads {
			w.cand = key
			if now.Sub(w.lastAt) >= rankSettle {
				return w.finish()
			}
		}
	}
	switch {
	case !now.Before(w.started.Add(rankWatchMax)):
		return w.finish()
	case w.screenUp(now):
		return RankRead{}, false
	case len(w.votes) > 0 && !w.seenAt.IsZero():
		return w.finish() // the screen closed
	case !now.Before(w.started.Add(rankSearchFor)):
		return w.finish() // never showed (or only unreadable)
	}
	return RankRead{}, false
}

// finish commits the latest confirmed read, else the latest expected one, else
// the most seen one (the latest on a tie), if any.
func (w *rankWatch) finish() (RankRead, bool) {
	best, n := w.cand, 0
	if best.Role == "" {
		best = w.exp
	}
	if best.Role == "" {
		for _, k := range w.order {
			if w.votes[k] >= n {
				best, n = k, w.votes[k]
			}
		}
	}
	w.stop()
	return best, best.Role != ""
}
