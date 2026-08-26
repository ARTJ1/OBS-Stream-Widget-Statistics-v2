package owtracker

import "testing"

func TestParseBannerOutcomeWinRU(t *testing.T) {
	if parseBannerOutcome("П О Б Е Д А !") != "win" {
		t.Fatal("expected win for ПОБЕДА")
	}
}

func TestParseBannerOutcomeLossRU(t *testing.T) {
	if parseBannerOutcome("ПОРАЖЕНИЕ") != "loss" {
		t.Fatal("expected loss for ПОРАЖЕНИЕ")
	}
	if parseBannerOutcome("ПИРИЖЕНИЕ") != "loss" {
		t.Fatal("expected loss for common OCR typo ПИРИЖЕНИЕ")
	}
}

func TestParseBannerOutcomeWinEN(t *testing.T) {
	if parseBannerOutcome("VICTORY") != "win" {
		t.Fatal("expected win for VICTORY")
	}
}

func TestParseBannerOutcomeGameplayText(t *testing.T) {
	if parseBannerOutcome("ЗАЩИЩАЙТЕ") != "" {
		t.Fatal("gameplay text must not trigger")
	}
}

func TestConfirmOCRConsensus(t *testing.T) {
	a := ProbeResult{Match: "win", MatchMethod: "ocr", WouldTrigger: true, OcrText: "VICTORY"}
	b := ProbeResult{Match: "win", MatchMethod: "ocr", WouldTrigger: true, OcrText: "VICTORY!"}
	c := ProbeResult{Match: "loss", MatchMethod: "ocr", WouldTrigger: true}
	confirmed, ok := confirmOCRConsensus([]ProbeResult{a, b, c})
	if ok || confirmed.Match == "win" {
		t.Fatalf("expected reject mixed OCR, ok=%v match=%q", ok, confirmed.Match)
	}
	confirmed, ok = confirmOCRConsensus([]ProbeResult{a, b})
	if !ok || confirmed.Match != "win" {
		t.Fatalf("expected win consensus, ok=%v match=%q", ok, confirmed.Match)
	}
}
