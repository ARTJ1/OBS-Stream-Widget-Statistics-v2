package owtracker

import "image"

const (
	bannerShapeMinScore = 80
	applyColorMin       = 15
	applyColorGap       = 10
)

// bannerShapeVisible reports a wide gold/red banner word band (language-independent).
func bannerShapeVisible(crop image.Image, goldPct, defeatPct int) bool {
	if crop == nil {
		return false
	}
	if endScreenBannerImage(crop) {
		return true
	}
	band, ok := extractBannerTextBand(crop, goldPct, defeatPct)
	if !ok {
		return false
	}
	return bannerBinaryScore(bannerBinary(band)) >= bannerShapeMinScore
}

func applyColorStrong(label string, gold, defeat int) bool {
	switch label {
	case "win":
		return gold >= applyColorMin && gold > defeat+applyColorGap
	case "loss":
		return defeat >= applyColorMin && defeat > gold+applyColorGap
	default:
		return false
	}
}

// bannerShapeOutcome requires banner glyph band + strong palette for the outcome.
func bannerShapeOutcome(crop image.Image, r ProbeResult) (string, bool) {
	if !r.EndScreenActive && !r.HashEndScreen {
		return "", false
	}
	if !bannerShapeVisible(crop, r.GoldPct, r.DefeatPct) {
		return "", false
	}
	if applyColorStrong("win", r.GoldPct, r.DefeatPct) {
		return "win", true
	}
	if applyColorStrong("loss", r.GoldPct, r.DefeatPct) {
		return "loss", true
	}
	return "", false
}

func applyProbeAllowed(probe ProbeResult) bool {
	if probe.Match != "win" && probe.Match != "loss" {
		return false
	}
	if probe.MatchMethod == "text" {
		return probe.EndScreenActive || probe.Source == "preview" || probe.Source == "upload"
	}
	if !probe.EndScreenActive && !probe.HashEndScreen {
		return false
	}
	if probe.MatchMethod == "ml" {
		return applyColorStrong(probe.Match, probe.GoldPct, probe.DefeatPct)
	}
	return applyColorStrong(probe.Match, probe.GoldPct, probe.DefeatPct)
}
