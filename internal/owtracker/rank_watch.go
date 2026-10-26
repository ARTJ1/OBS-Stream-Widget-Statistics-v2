package owtracker

import "time"

// rankWatch turns per-frame progress-screen reads into at most one rank per match.
//
//   - it starts when a win/loss is counted and runs for rankWatchFor (the screen
//     shows after Play of the Game, ~50 s later, or at once if the player leaves)
//   - a rank is confirmed by rankConfirmReads equal reads, or by a single read once
//     the screen is gone (players who click Continue fast leave it up ~2 s)
//   - the label already shows the new rank while the emblem animates, so reads
//     during a rank-up agree with each other
const (
	rankWatchFor     = 150 * time.Second
	rankConfirmReads = 2
	rankGoneFrames   = 3
	rankPoll         = 250 * time.Millisecond // one frame per rankPoll, work included
	rankMinPause     = 40 * time.Millisecond  // breather when a read takes longer
)

// Outside the post-match watch (rank_gate.go): the gate opens a short watch.
const (
	rankGateWatchFor  = 8 * time.Second         // the screen it found shows its rank at once
	rankGateQuietFor  = 60 * time.Second        // after a gate watch: the same screen again
	rankGateMaxMisses = 3                       // gate passes with no role header read …
	rankGateBackoff   = 1500 * time.Millisecond // … then at most one read this often
)

type rankWatch struct {
	until time.Time
	votes map[RankRead]int
	order []RankRead // first-seen order, for ties
	gone  int
}

func (w *rankWatch) start(now time.Time) { w.startFor(now, rankWatchFor) }

func (w *rankWatch) startFor(now time.Time, d time.Duration) {
	*w = rankWatch{until: now.Add(d), votes: map[RankRead]int{}}
}

// active stays true until observe has run past the deadline (so a lone read is
// still committed).
func (w *rankWatch) active() bool { return w.votes != nil }

func (w *rankWatch) stop() { *w = rankWatch{} }

// observe returns the confirmed rank exactly once per watch.
func (w *rankWatch) observe(now time.Time, r RankRead) (RankRead, bool) {
	if w.votes == nil {
		return RankRead{}, false
	}
	if !r.Valid() {
		if r.Role != "" {
			w.gone = 0 // screen still up, label mid-animation
		} else if len(w.votes) > 0 {
			w.gone++
			if w.gone >= rankGoneFrames {
				return w.finish()
			}
		}
		if !now.Before(w.until) {
			return w.finish()
		}
		return RankRead{}, false
	}
	w.gone = 0
	key := RankRead{Role: r.Role, Tier: r.Tier, Division: r.Division, Rank: r.Rank}
	if w.votes[key] == 0 {
		w.order = append(w.order, key)
	}
	w.votes[key]++
	if w.votes[key] >= rankConfirmReads {
		w.stop()
		return key, true
	}
	return RankRead{}, false
}

// finish commits the most seen read (the latest one on a tie), if any.
func (w *rankWatch) finish() (RankRead, bool) {
	best, n := RankRead{}, 0
	for _, k := range w.order {
		if w.votes[k] >= n {
			best, n = k, w.votes[k]
		}
	}
	w.stop()
	return best, n > 0
}
