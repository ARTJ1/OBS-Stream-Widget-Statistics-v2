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
const (
	rankSearchFor    = flowMinMatchGap
	rankWatchMax     = 3 * time.Minute
	rankConfirmReads = 2
	rankSettle       = 15 * time.Second
	rankGoneAfter    = 2 * time.Second
	rankPoll         = 250 * time.Millisecond // one frame per rankPoll, work included
	rankMinPause     = 40 * time.Millisecond  // breather when a read takes longer
)

type rankWatch struct {
	started time.Time
	seenAt  time.Time // last frame with the role read (screen up)
	votes   map[RankRead]int
	order   []RankRead // first-seen order, for ties
	last    RankRead   // latest valid read and how many times in a row
	lastN   int
	lastAt  time.Time
	cand    RankRead // latest read confirmed rankConfirmReads times in a row
}

func (w *rankWatch) start(now time.Time) {
	*w = rankWatch{started: now, votes: map[RankRead]int{}}
}

// active stays true until observe has run past the end (so a lone read is
// still committed).
func (w *rankWatch) active() bool { return w.votes != nil }

func (w *rankWatch) stop() { *w = rankWatch{} }

// screenUp tells whether the progress screen was read within rankGoneAfter.
func (w *rankWatch) screenUp(now time.Time) bool {
	return !w.seenAt.IsZero() && now.Sub(w.seenAt) < rankGoneAfter
}

// observe returns the confirmed rank exactly once per watch.
func (w *rankWatch) observe(now time.Time, r RankRead) (RankRead, bool) {
	if w.votes == nil {
		return RankRead{}, false
	}
	if r.Role != "" {
		w.seenAt = now
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

// finish commits the latest confirmed read, else the most seen one (the latest
// on a tie), if any.
func (w *rankWatch) finish() (RankRead, bool) {
	best, n := w.cand, 0
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
