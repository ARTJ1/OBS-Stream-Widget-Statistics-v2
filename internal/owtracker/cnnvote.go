package owtracker

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"time"
)

// Banner CNN decision rule (frames arrive ~4 per second):
//   - a frame votes win/loss when that class has p >= cnnFrameP
//   - a result is reported only after cnnRun frames IN A ROW vote the same with
//     p >= cnnStrongP, and no frame voted the opposite result (p >= cnnFrameP) in
//     the last cnnOppositeWindow (the victory fly-in flash can look like defeat
//     for a frame; this turns it into a short delay instead of a wrong count)
//   - autoFlow then requires two such reports in a row (~1 s of agreeing frames)
//     and counts once per match. Measured on a 6 h stream: stricter beats faster —
//     a wrong count is worse than a miss (the streamer still has the hotkey).
const (
	cnnFrameP         = 0.5
	cnnStrongP        = 0.85
	cnnRun            = 3
	cnnOppositeWindow = 2 * time.Second
	cnnRunMaxSpan     = 1500 * time.Millisecond
)

//go:embed assets/models/banner_cnn.json.gz
var embeddedBannerCNN []byte

func loadEmbeddedBannerCNN() (*BannerCNN, error) {
	zr, err := gzip.NewReader(bytes.NewReader(embeddedBannerCNN))
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return parseBannerCNN(b)
}

type cnnFrame struct {
	at    time.Time
	label Outcome // "" = no vote
	p     float32
}

type cnnVote struct {
	frames []cnnFrame
}

func (v *cnnVote) reset() { v.frames = nil }

// add records one frame and returns what the flow should see for it.
func (v *cnnVote) add(at time.Time, win, loss float32) BannerRead {
	f := cnnFrame{at: at}
	switch {
	case win >= loss && win >= cnnFrameP:
		f.label, f.p = OutcomeWin, win
	case loss > win && loss >= cnnFrameP:
		f.label, f.p = OutcomeLoss, loss
	}
	v.frames = append(v.frames, f)
	cut := 0
	for cut < len(v.frames) && at.Sub(v.frames[cut].at) > cnnOppositeWindow {
		cut++
	}
	v.frames = v.frames[cut:]

	r := BannerRead{Likely: f.label != "", WinP: win, LossP: loss, Variant: -1}
	n := len(v.frames)
	if f.label == "" || n < cnnRun {
		return r
	}
	run := v.frames[n-cnnRun:]
	if at.Sub(run[0].at) > cnnRunMaxSpan {
		return r
	}
	for _, rf := range run {
		if rf.label != f.label || rf.p < cnnStrongP {
			return r
		}
	}
	for _, of := range v.frames {
		if of.label != "" && of.label != f.label {
			return r
		}
	}
	r.Outcome = f.label
	return r
}
