package owtracker

import (
	"testing"
	"time"
)

// feed simulates frames 250 ms apart; returns the first counted outcome and when.
func feed(t *testing.T, frames [][2]float32) (Outcome, int) {
	t.Helper()
	var v cnnVote
	var f autoFlow
	t0 := time.Date(2026, 10, 5, 20, 0, 0, 0, time.UTC)
	for i, p := range frames {
		at := t0.Add(time.Duration(i) * 250 * time.Millisecond)
		if o, ok := f.observe(at, v.add(at, p[0], p[1])); ok {
			return o, i
		}
	}
	return "", -1
}

var (
	none = [2]float32{0, 0}
	win  = [2]float32{0.98, 0}
	loss = [2]float32{0, 0.98}
	weak = [2]float32{0.7, 0}
)

func rep(f [2]float32, n int) [][2]float32 {
	out := make([][2]float32, n)
	for i := range out {
		out[i] = f
	}
	return out
}

func cat(parts ...[][2]float32) [][2]float32 {
	var out [][2]float32
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func TestVoteCountsSteadyBanner(t *testing.T) {
	o, i := feed(t, cat(rep(none, 8), rep(loss, 12), rep(none, 8)))
	if o != OutcomeLoss {
		t.Fatalf("got %q", o)
	}
	if i > 8+4 {
		t.Fatalf("too slow: counted at frame %d (banner started at 8)", i)
	}
}

func TestVoteShortDimBanner(t *testing.T) {
	// 1.5 s banner (6 frames), like the shortest one in the test stream.
	if o, _ := feed(t, cat(rep(none, 4), rep(win, 6), rep(none, 4))); o != OutcomeWin {
		t.Fatalf("short banner missed: %q", o)
	}
}

func TestVoteIgnoresIsolatedFalseFrames(t *testing.T) {
	seq := cat(rep(none, 10), rep(win, 1), rep(none, 10), rep(loss, 2), rep(none, 10), rep(weak, 6), rep(none, 4))
	if o, _ := feed(t, seq); o != "" {
		t.Fatalf("false count %q", o)
	}
}

func TestVoteVictoryFlashNeverCountsDefeat(t *testing.T) {
	// Fly-in flash read as "loss" for 2 frames, then the real victory banner.
	o, i := feed(t, cat(rep(none, 4), rep(loss, 2), rep(win, 14), rep(none, 4)))
	if o != OutcomeWin {
		t.Fatalf("got %q, want win", o)
	}
	t.Logf("victory counted at frame %d (flash at 4-5)", i)
}

func TestVoteMixedFramesDoNotCount(t *testing.T) {
	seq := cat(rep(none, 4))
	for i := 0; i < 10; i++ {
		seq = append(seq, win, loss)
	}
	if o, _ := feed(t, seq); o != "" {
		t.Fatalf("alternating win/loss counted %q", o)
	}
}

func TestVoteFlashThenShortVictory(t *testing.T) {
	// Worst case: flash read as defeat, then only 1.5 s of victory. The strict
	// rule may miss it (streamer presses the hotkey) but must never count defeat.
	if o, _ := feed(t, cat(rep(none, 4), rep(loss, 2), rep(win, 6), rep(none, 4))); o == OutcomeLoss {
		t.Fatalf("counted defeat for a victory")
	}
}

func TestVoteDarkTransitionDoesNotCount(t *testing.T) {
	// Dark round-transition frames read as weak "loss" for ~1 s (4:48:59 in the test stream).
	seq := cat(rep(none, 8), rep([2]float32{0, 0.9}, 3), rep([2]float32{0, 0.6}, 2), rep(none, 8))
	if o, _ := feed(t, seq); o != "" {
		t.Fatalf("false count %q", o)
	}
}
