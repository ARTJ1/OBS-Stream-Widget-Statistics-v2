package owtracker

import (
	"image/color"
	"testing"
)

func TestEndScreenGateBlocksGameplayLoss(t *testing.T) {
	// Log case: mid-game HUD, loss hash closer (d=20 vs 36, 69% similarity) but no defeat banner.
	img := solid(1001, 297, color.RGBA{R: 40, G: 35, B: 50, A: 255})
	match, ok, _, method := pickMatchScores(36, 20, 2, 0, 8, img)
	if ok || match == "loss" {
		t.Fatalf("expected no loss on gameplay scene, got match=%q ok=%v method=%s", match, ok, method)
	}
}

func TestEndScreenGateAllowsDefeatBanner(t *testing.T) {
	img := solid(480, 180, color.RGBA{R: 200, G: 40, B: 40, A: 255})
	match, ok, _, method := pickMatchScores(36, 18, 10, 5, 8, img)
	if !ok || match != "loss" {
		t.Fatalf("expected loss on red defeat banner, got match=%q ok=%v method=%s", match, ok, method)
	}
}

func TestConfirmProbeConsensusColorVictory(t *testing.T) {
	first := ProbeResult{
		Match: "win", MatchMethod: "color-victory", WouldTrigger: true,
		GoldPct: 21, DefeatPct: 5,
	}
	second := ProbeResult{
		Match: "", WouldTrigger: false,
		GoldPct: 14, DefeatPct: 5,
	}
	third := ProbeResult{
		Match: "", WouldTrigger: false,
		GoldPct: 12, DefeatPct: 8,
	}
	confirmed, ok := confirmProbeConsensus([]ProbeResult{first, second, third})
	if !ok || confirmed.Match != "win" {
		t.Fatalf("expected color-victory consensus win, got ok=%v match=%q", ok, confirmed.Match)
	}
}

func TestAllowsLossTriggerRejectsWeakHash(t *testing.T) {
	if allowsLossTrigger(2, 2, 0, 0, 20, 8) {
		t.Fatal("expected weak gameplay hash to be rejected")
	}
}

func TestAllowsWinTriggerAcceptsGold(t *testing.T) {
	if !allowsWinTrigger(12, 3, 10, 8, 30, 8) {
		t.Fatal("expected gold banner to allow win")
	}
}
