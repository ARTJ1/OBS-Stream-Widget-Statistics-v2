package owtracker

import (
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

func TestMLColorAllowsWinGold(t *testing.T) {
	if !mlColorAllows("win", 28, 8) {
		t.Fatal("gold banner should allow win")
	}
	if mlColorAllows("win", 12, 10) {
		t.Fatal("weak gold gap should not allow win")
	}
}

func TestMLHashVetoLossOnGoldBanner(t *testing.T) {
	r := ProbeResult{
		GoldPct:        28,
		DefeatPct:      6,
		WinSimilarity:  62,
		LossSimilarity: 40,
	}
	if !mlHashVeto(r, "loss") {
		t.Fatal("loss should be vetoed on gold banner")
	}
}

func TestMLMarginAllows(t *testing.T) {
	ok := mlMarginAllows(ml.Result{
		Label:    ml.LabelWin,
		WinProb:  0.85,
		LossProb: 0.10,
		NoneProb: 0.05,
	})
	if !ok {
		t.Fatal("expected margin ok")
	}
	reject := mlMarginAllows(ml.Result{
		Label:    ml.LabelLoss,
		WinProb:  0.55,
		LossProb: 0.40,
		NoneProb: 0.05,
	})
	if reject {
		t.Fatal("ambiguous loss should fail margin")
	}
}
