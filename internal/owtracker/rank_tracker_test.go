package owtracker

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
	"time"
)

// rankFlowTracker is a Tracker with an OBS frame source and a scripted reader:
// the live auto rank path without Windows OCR or a game.
type rankFlowTracker struct {
	*Tracker
	script  []RankRead // one per reader call; empty read after the end
	calls   int
	widths  []int
	applied []RankRead
}

func newRankFlowTracker(t *testing.T, rankOn bool, script ...RankRead) *rankFlowTracker {
	t.Helper()
	ft := &rankFlowTracker{Tracker: &Tracker{dataDir: t.TempDir()}, script: script}
	ft.ocr = newWinOCR(ft.dataDir)
	ft.rankOn.Store(rankOn)
	ft.SetFrameSource(func(width int) (image.Image, error) {
		ft.widths = append(ft.widths, width)
		return image.NewRGBA(image.Rect(0, 0, width, width*9/16)), nil
	})
	ft.rankReader = func(image.Image) (RankRead, error) {
		ft.calls++
		if ft.calls <= len(ft.script) {
			return ft.script[ft.calls-1], nil
		}
		return RankRead{}, nil
	}
	ft.SetRankHandler(func(r RankRead) { ft.applied = append(ft.applied, r) })
	return ft
}

// countMatch feeds the banner reads that count one win (and start the watch).
func (ft *rankFlowTracker) countMatch() {
	for i := 0; i < flowConfirmFrames; i++ {
		ft.handleRead(BannerRead{Likely: true, Outcome: OutcomeWin})
	}
}

// endCooldown moves the last count back so the next tick is past the cooldown.
func (ft *rankFlowTracker) endCooldown() {
	ft.autoMu.Lock()
	ft.flow.lastCount = time.Now().Add(-flowMinMatchGap - time.Second)
	ft.flow.lastBanner = ft.flow.lastCount
	ft.autoMu.Unlock()
}

var (
	em2      = RankRead{Role: "damage", Tier: "Emerald", Division: 2, Rank: RankIndex("Emerald", 2)}
	roleOnly = RankRead{Role: "damage"}
)

func TestAutoRankFlowConfirms(t *testing.T) {
	ft := newRankFlowTracker(t, true, RankRead{}, roleOnly, em2, em2)
	ft.countMatch()
	if !ft.rank.active() || !ft.Status().RankWatch {
		t.Fatal("watch not started after a counted match")
	}
	for i := 0; i < 10 && len(ft.applied) == 0; i++ {
		if d := ft.autoTick(false); d < rankMinPause || d > rankPoll {
			t.Fatalf("tick %d during the watch returned %s, want up to %s", i, d, rankPoll)
		}
	}
	if len(ft.applied) != 1 || ft.applied[0] != em2 {
		t.Fatalf("applied %+v, want one %+v", ft.applied, em2)
	}
	if ft.calls != 4 {
		t.Errorf("reader calls = %d, want 4 (confirm on the 2nd equal read)", ft.calls)
	}
	for _, w := range ft.widths {
		if w != rankFrameWidth {
			t.Errorf("asked OBS for %d px, want %d", w, rankFrameWidth)
		}
	}
	st := ft.Status()
	if st.RankWatch || st.LastRank == nil || st.LastRank.RankRead != em2 {
		t.Errorf("status after confirm: watch=%v last=%+v", st.RankWatch, st.LastRank)
	}
	// The rest of the cooldown captures nothing.
	calls := ft.calls
	if d := ft.autoTick(false); d != 5*time.Second || ft.calls != calls {
		t.Errorf("after confirm: delay %s, reader calls %d -> %d", d, calls, ft.calls)
	}
}

func TestAutoRankFlowOff(t *testing.T) {
	ft := newRankFlowTracker(t, false, em2, em2)
	ft.countMatch()
	if ft.rank.active() {
		t.Fatal("watch started with auto rank off")
	}
	for i := 0; i < 5; i++ {
		ft.autoTick(false)
	}
	if ft.calls != 0 || len(ft.widths) != 0 || len(ft.applied) != 0 {
		t.Errorf("auto rank off: reader %d, frames %d, applied %d", ft.calls, len(ft.widths), len(ft.applied))
	}
}

func TestAutoRankFlowDryRun(t *testing.T) {
	ft := newRankFlowTracker(t, true, em2, em2)
	ft.dryRun.Store(true)
	ft.countMatch()
	for i := 0; i < 5; i++ {
		ft.autoTick(false)
	}
	if len(ft.applied) != 0 {
		t.Errorf("dry run applied %+v", ft.applied)
	}
	if st := ft.Status(); st.LastRank == nil || st.LastRank.RankRead != em2 {
		t.Errorf("dry run should still show the read, got %+v", st.LastRank)
	}
}

func TestAutoRankFlowLoneReadAtCooldownEnd(t *testing.T) {
	// One good read, then the screen is still "there" (role only) until the cooldown ends.
	ft := newRankFlowTracker(t, true, em2, roleOnly, roleOnly)
	ft.countMatch()
	for i := 0; i < 3; i++ {
		ft.autoTick(false)
	}
	if len(ft.applied) != 0 {
		t.Fatalf("applied before the cooldown ended: %+v", ft.applied)
	}
	ft.endCooldown()
	ft.autoTick(false)
	if len(ft.applied) != 1 || ft.applied[0] != em2 {
		t.Fatalf("lone read not committed at cooldown end: %+v", ft.applied)
	}
	if ft.rank.active() || ft.Status().RankWatch {
		t.Error("watch still active after the cooldown")
	}
}

func TestAutoRankFlowUnreadableLabel(t *testing.T) {
	// The screen shows, but the label never reads: the panel gets what OCR saw.
	seen := RankRead{Role: "damage", RoleText: "DAMAGE", RankText: "EMEPALO ?"}
	ft := newRankFlowTracker(t, true, seen, seen)
	ft.countMatch()
	for i := 0; i < 3; i++ {
		ft.autoTick(false)
	}
	ft.endCooldown()
	ft.autoTick(false)
	if len(ft.applied) != 0 {
		t.Fatalf("applied %+v", ft.applied)
	}
	st := ft.Status()
	if st.RankMiss == nil || st.RankMiss.RankText != seen.RankText {
		t.Fatalf("miss not reported: %+v", st.RankMiss)
	}
	// The next match reads fine: the miss is cleared.
	ft.script, ft.calls = []RankRead{em2, em2}, 0
	ft.flow.reset()
	ft.countMatch()
	for i := 0; i < 3; i++ {
		ft.autoTick(false)
	}
	if st := ft.Status(); st.RankMiss != nil || st.LastRank == nil {
		t.Errorf("after a good read: miss=%+v last=%+v", st.RankMiss, st.LastRank)
	}
}

func TestAutoRankFlowSwitchedOffMidWatch(t *testing.T) {
	ft := newRankFlowTracker(t, true, em2)
	ft.countMatch()
	ft.autoTick(false) // one read, not confirmed yet
	ft.SetRankAuto(false)
	ft.autoTick(false)
	ft.endCooldown()
	ft.autoTick(false)
	if len(ft.applied) != 0 {
		t.Errorf("applied after switching off: %+v", ft.applied)
	}
	if ft.rank.active() {
		t.Error("watch still active")
	}
}

func TestAutoRankConfigPersisted(t *testing.T) {
	dir := t.TempDir()
	tr := &Tracker{dataDir: dir, ocr: newWinOCR(dir)}
	tr.enabled.Store(true)
	tr.SetRankAuto(true)
	if cfg := loadAutoConfig(dir); !cfg.Enabled || !cfg.Rank {
		t.Fatalf("saved %+v", cfg)
	}
	tr.SetEnabled(false)
	if cfg := loadAutoConfig(dir); cfg.Enabled || !cfg.Rank {
		t.Fatalf("auto W/L off must keep the rank switch: %+v", cfg)
	}
}

// progressLikeFrame is a dark frame with a short light word where the progress
// screen has its role header: it passes the gate.
func progressLikeFrame(width int) image.Image {
	h := width * 9 / 16
	img := image.NewRGBA(image.Rect(0, 0, width, h))
	r := rankGateRect(width, h)
	cx := r.Min.X + r.Dx()/2
	for y := r.Min.Y + r.Dy()/3; y < r.Max.Y-r.Dy()/4; y++ {
		for x := cx - width/30; x < cx+width/30; x += 2 {
			img.Set(x, y, color.White)
		}
	}
	return img
}

func newGateFlowTracker(t *testing.T, script ...RankRead) *rankFlowTracker {
	t.Helper()
	ft := newRankFlowTracker(t, true, script...)
	m, err := loadEmbeddedBannerCNN()
	if err != nil {
		t.Skipf("banner model: %v", err)
	}
	ft.cnn = m
	ft.SetFrameSource(func(width int) (image.Image, error) {
		ft.widths = append(ft.widths, width)
		return progressLikeFrame(width), nil
	})
	return ft
}

func TestRankGate(t *testing.T) {
	if !rankGate(cropRankGate(progressLikeFrame(480))) {
		t.Error("progress-like frame did not pass")
	}
	if rankGate(cropRankGate(image.NewRGBA(image.Rect(0, 0, 480, 270)))) {
		t.Error("black frame passed")
	}
	white := image.NewRGBA(image.Rect(0, 0, 480, 270))
	draw.Draw(white, white.Bounds(), image.White, image.Point{}, draw.Src)
	if rankGate(cropRankGate(white)) {
		t.Error("white frame passed")
	}
}

func TestAutoRankGateWithoutCount(t *testing.T) {
	// No banner counted (a draw, a missed banner): the gate finds the screen.
	ft := newGateFlowTracker(t, em2, em2)
	for i := 0; i < 5 && len(ft.applied) == 0; i++ {
		ft.autoTick(false)
	}
	if len(ft.applied) != 1 || ft.applied[0] != em2 {
		t.Fatalf("applied %+v, want %+v", ft.applied, em2)
	}
	if ft.rank.active() || ft.rankByGate {
		t.Error("gate watch still open after the confirm")
	}
	// The same screen stays up: quiet, no new reads.
	calls := ft.calls
	for i := 0; i < 5; i++ {
		ft.autoTick(false)
	}
	if ft.calls != calls || len(ft.applied) != 1 {
		t.Errorf("read again during the quiet time: calls %d -> %d, applied %d", calls, ft.calls, len(ft.applied))
	}
}

func TestAutoRankGateRoleOnlyIsQuiet(t *testing.T) {
	// The gate read a role word but never a rank (gameplay text): the short watch
	// ends and the panel shows no "unreadable" note.
	ft := newGateFlowTracker(t, roleOnly, roleOnly, roleOnly)
	ft.autoTick(false)
	if !ft.rank.active() || !ft.rankByGate {
		t.Fatal("gate watch not opened")
	}
	ft.rank.until = time.Now() // deadline reached
	ft.autoTick(false)
	if ft.rank.active() || ft.rankByGate {
		t.Fatal("gate watch still open after its deadline")
	}
	if st := ft.Status(); st.RankMiss != nil || len(ft.applied) != 0 {
		t.Errorf("gate watch reported %+v, applied %+v", st.RankMiss, ft.applied)
	}
}

func TestAutoRankGateBacksOff(t *testing.T) {
	// The gate passes on something that is not the progress screen.
	ft := newGateFlowTracker(t)
	for i := 0; i < 20; i++ {
		ft.autoTick(false)
	}
	if ft.calls != rankGateMaxMisses {
		t.Errorf("reader calls = %d, want %d then a back-off", ft.calls, rankGateMaxMisses)
	}
	if ft.rank.active() || len(ft.applied) != 0 {
		t.Error("watch opened without a role read")
	}
}
