package owtracker

import (
	"strings"
)

func confirmOCRConsensus(probes []ProbeResult) (ProbeResult, bool) {
	if len(probes) < 2 {
		return ProbeResult{}, false
	}
	match := ""
	agree := 0
	var first ProbeResult
	for _, p := range probes {
		if !p.WouldTrigger || p.Match == "" {
			continue
		}
		if p.MatchMethod != "ocr" {
			continue
		}
		if match == "" {
			match = p.Match
			first = p
		}
		if p.Match != match {
			return ProbeResult{}, false
		}
		agree++
	}
	if agree >= 2 {
		return first, true
	}
	return ProbeResult{}, false
}

var (
	winKeywords = []string{
		"VICTORY", "VICTOR", "ПОБЕДА", "ПОБЕД", "POBEDA", "POBED",
	}
	lossKeywords = []string{
		"DEFEAT", "DEFEA", "ПОРАЖЕНИЕ", "ПОРАЖЕН", "ПИРИЖЕНИЕ", "PORAZHENIE", "PORAZHEN",
	}
)

// parseBannerOutcome maps OCR text to win/loss. Empty string means no confident match.
func parseBannerOutcome(raw string) string {
	norm := normalizeOCRText(raw)
	if norm == "" {
		return ""
	}
	winHit := fuzzyKeywordHit(norm, winKeywords)
	lossHit := fuzzyKeywordHit(norm, lossKeywords)
	if winHit && !lossHit {
		return "win"
	}
	if lossHit && !winHit {
		return "loss"
	}
	return ""
}

func containsKeyword(norm string, keys []string) bool {
	for _, k := range keys {
		if strings.Contains(norm, k) {
			return true
		}
	}
	return false
}
