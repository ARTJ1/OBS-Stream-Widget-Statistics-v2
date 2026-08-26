package owtracker

const (
	bannerColorMin   = 18 // % gold or red pixels — banner must be visible
	bannerColorGap   = 6 // win gold must exceed defeat by this much (and vice versa)
	bannerHashMinSim = 50 // % pHash similarity minimum
)

// strictBannerOutcome scores only when banner colors are strong AND hash agrees.
// No loose template guessing — prevents false +loss on gameplay UI.
func strictBannerOutcome(r ProbeResult) (string, bool) {
	if !r.HashEndScreen && !r.EndScreenActive {
		return "", false
	}
	gold, defeat := r.GoldPct, r.DefeatPct
	winSim, lossSim := r.WinSimilarity, r.LossSimilarity

	hasGoldBanner := gold >= bannerColorMin && gold > defeat+bannerColorGap
	hasDefeatBanner := defeat >= bannerColorMin && defeat > gold+bannerColorGap

	if hasGoldBanner && winSim >= bannerHashMinSim && winSim >= lossSim-8 {
		return "win", true
	}
	if hasDefeatBanner && lossSim >= bannerHashMinSim && lossSim >= winSim-8 {
		return "loss", true
	}
	return "", false
}

func bannerColorsVisible(gold, defeat int) bool {
	return gold >= 12 || defeat >= 12
}
