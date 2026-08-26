package owtracker

import "testing"

func TestStrictBannerVictory(t *testing.T) {
	r := ProbeResult{
		HashEndScreen: true, EndScreenActive: true,
		GoldPct: 21, DefeatPct: 14,
		WinSimilarity: 59, LossSimilarity: 66,
	}
	match, ok := strictBannerOutcome(r)
	if !ok || match != "win" {
		t.Fatalf("expected win on gold banner, got ok=%v match=%q", ok, match)
	}
}

func TestStrictBannerRejectsWeakColor(t *testing.T) {
	r := ProbeResult{
		HashEndScreen: true, EndScreenActive: true,
		DefeatPct: 4, GoldPct: 0,
		WinSimilarity: 47, LossSimilarity: 59,
	}
	if _, ok := strictBannerOutcome(r); ok {
		t.Fatal("defeatPct=4 must not trigger loss")
	}
}

func TestStrictBannerRejectsEqualColors(t *testing.T) {
	r := ProbeResult{
		HashEndScreen: true, EndScreenActive: true,
		GoldPct: 6, DefeatPct: 6,
		WinSimilarity: 41, LossSimilarity: 53,
	}
	if _, ok := strictBannerOutcome(r); ok {
		t.Fatal("equal weak colors must not trigger")
	}
}

func TestStrictBannerRejectsExplorer(t *testing.T) {
	r := ProbeResult{
		HashEndScreen: true, EndScreenActive: true,
		GoldPct: 0, DefeatPct: 0,
		WinSimilarity: 56, LossSimilarity: 47,
	}
	if _, ok := strictBannerOutcome(r); ok {
		t.Fatal("no banner colors must not trigger")
	}
}
