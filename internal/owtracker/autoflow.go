package owtracker

import "time"

// autoFlow turns per-frame banner reads into at most one result per match.
//
//   - a result needs flowConfirmFrames reads with the same word within flowConfirmWindow
//   - after counting, the flow is disarmed: it re-arms only once no banner has been
//     visible for flowRearmGap AND flowMinMatchGap has passed (no match is shorter)
//   - a draw banner disarms without counting
const (
	flowConfirmFrames = 2
	flowConfirmWindow = 4 * time.Second
	flowRearmGap      = 15 * time.Second
	flowMinMatchGap   = 150 * time.Second
)

type autoFlow struct {
	disarmed   bool
	lastCount  time.Time
	lastBanner time.Time
	cand       Outcome
	candN      int
	candAt     time.Time
	// need overrides flowConfirmFrames (the banner CNN vote already requires
	// several frames in a row, so its reports count on the first one).
	need int
}

func (f *autoFlow) armed() bool { return !f.disarmed }

func (f *autoFlow) reset() { *f = autoFlow{need: f.need} }

// observe returns (win|loss, true) exactly once per match.
func (f *autoFlow) observe(now time.Time, r BannerRead) (Outcome, bool) {
	if r.Likely {
		f.lastBanner = now
	}
	if f.disarmed {
		if now.Sub(f.lastBanner) >= flowRearmGap && now.Sub(f.lastCount) >= flowMinMatchGap {
			f.disarmed = false
		}
		return "", false
	}
	if r.Outcome == "" {
		if f.cand != "" && now.Sub(f.candAt) > flowConfirmWindow {
			f.cand, f.candN = "", 0
		}
		return "", false
	}
	if r.Outcome == f.cand && now.Sub(f.candAt) <= flowConfirmWindow {
		f.candN++
	} else {
		f.cand, f.candN = r.Outcome, 1
	}
	f.candAt = now
	need := flowConfirmFrames
	if f.need > 0 {
		need = f.need
	}
	if f.candN < need {
		return "", false
	}
	outcome := f.cand
	f.cand, f.candN = "", 0
	f.disarmed = true
	f.lastCount = now
	if outcome == OutcomeDraw {
		return "", false
	}
	return outcome, true
}
