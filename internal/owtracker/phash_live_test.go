package owtracker

import "testing"

func TestConfirmsProbeMatchRejectsLossOnGold(t *testing.T) {
	first := ProbeResult{Match: "loss", GoldPct: 15, DefeatPct: 2, WinPixelPct: 44, LossPixelPct: 56, WouldTrigger: true}
	second := ProbeResult{Match: "loss", GoldPct: 14, DefeatPct: 2, WinPixelPct: 45, LossPixelPct: 55, WouldTrigger: true}
	if confirmsProbeMatch(first, second) {
		t.Fatal("expected reject loss when gold victory banner present")
	}
}
