package owtracker

import "testing"

func TestInstantApplyDisabled(t *testing.T) {
	p := ProbeResult{
		Match: "loss", MatchMethod: "banner", WouldTrigger: true,
		DefeatPct: 49,
	}
	if instantApply(p) {
		t.Fatal("instant apply must be disabled")
	}
}
