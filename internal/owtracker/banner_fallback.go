package owtracker

import "image"

// endScreenBannerFallback uses banner shape + colors (not hash/reference matching).
func endScreenBannerFallback(r ProbeResult, crop image.Image, _, _, _, _, _ int) (string, bool) {
	if match, ok := strictBannerOutcome(r); ok && applyColorStrong(match, r.GoldPct, r.DefeatPct) {
		if bannerShapeVisible(crop, r.GoldPct, r.DefeatPct) {
			return match, true
		}
	}
	return bannerShapeOutcome(crop, r)
}
