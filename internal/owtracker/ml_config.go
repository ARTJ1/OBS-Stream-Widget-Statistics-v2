package owtracker

const (
	mlMinConfidence      = 0.78 // default when only HUD/hash hint
	mlMinConfidenceEnd   = 0.52 // end-screen gate or strong banner colors
	mlMinConfidenceProbe = 0.45 // admin preview/upload (manual check)
)

func mlConfidenceThreshold(endActive, hashEnd bool, goldPct, defeatPct int, source string) float32 {
	if source == "preview" || source == "upload" {
		return mlMinConfidenceProbe
	}
	if endActive || hashEnd || bannerColorsVisible(goldPct, defeatPct) {
		return mlMinConfidenceEnd
	}
	return mlMinConfidence
}
