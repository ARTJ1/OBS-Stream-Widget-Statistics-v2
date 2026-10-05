package owtracker

import (
	"testing"
	"time"
)

func TestAutoFlowCountsOncePerMatch(t *testing.T) {
	var f autoFlow
	t0 := time.Date(2026, 10, 2, 20, 0, 0, 0, time.UTC)
	at := func(ms int) time.Time { return t0.Add(time.Duration(ms) * time.Millisecond) }
	win := BannerRead{Likely: true, Outcome: OutcomeWin}
	blank := BannerRead{}

	if _, ok := f.observe(at(0), win); ok {
		t.Fatal("one frame must not count")
	}
	o, ok := f.observe(at(700), win)
	if !ok || o != OutcomeWin {
		t.Fatalf("second frame should count, got %q %v", o, ok)
	}
	// Banner stays on screen for seconds: never counted again.
	for ms := 1400; ms < 9000; ms += 700 {
		if _, ok := f.observe(at(ms), win); ok {
			t.Fatalf("recounted at %dms", ms)
		}
	}
	// Banner gone, but too soon for a new match.
	if _, ok := f.observe(at(30_000), blank); ok || f.armed() {
		t.Fatal("re-armed too early")
	}
	// Next match ends minutes later.
	f.observe(at(200_000), blank)
	if !f.armed() {
		t.Fatal("should re-arm after banner absence and min match gap")
	}
	f.observe(at(400_000), BannerRead{Likely: true, Outcome: OutcomeLoss})
	o, ok = f.observe(at(400_700), BannerRead{Likely: true, Outcome: OutcomeLoss})
	if !ok || o != OutcomeLoss {
		t.Fatalf("next match loss: %q %v", o, ok)
	}
}

func TestAutoFlowNeedsAgreeingFrames(t *testing.T) {
	var f autoFlow
	t0 := time.Now()
	f.observe(t0, BannerRead{Likely: true, Outcome: OutcomeWin})
	if _, ok := f.observe(t0.Add(500*time.Millisecond), BannerRead{Likely: true, Outcome: OutcomeLoss}); ok {
		t.Fatal("disagreeing frames counted")
	}
	// Frames too far apart do not confirm each other.
	f.reset()
	f.observe(t0, BannerRead{Likely: true, Outcome: OutcomeWin})
	if _, ok := f.observe(t0.Add(flowConfirmWindow+time.Second), BannerRead{Likely: true, Outcome: OutcomeWin}); ok {
		t.Fatal("stale frame confirmed")
	}
}

func TestAutoFlowDrawDisarms(t *testing.T) {
	var f autoFlow
	t0 := time.Now()
	draw := BannerRead{Likely: true, Outcome: OutcomeDraw}
	f.observe(t0, draw)
	if _, ok := f.observe(t0.Add(time.Second), draw); ok {
		t.Fatal("draw counted")
	}
	if f.armed() {
		t.Fatal("draw should disarm")
	}
}
