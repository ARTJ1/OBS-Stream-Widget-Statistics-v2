package owtracker

import (
	"image"

	"github.com/nfnt/resize"
)

// pixelSimilarityScaled compares live crop to a reference template, resizing the
// reference when dimensions differ (Twitch crop size vs saved template crop).
func pixelSimilarityScaled(live, reference image.Image) int {
	if live == nil || reference == nil {
		return -1
	}
	lb := live.Bounds()
	if lb.Dx() < 8 || lb.Dy() < 8 {
		return -1
	}
	ref := reference
	rb := reference.Bounds()
	if rb.Dx() != lb.Dx() || rb.Dy() != lb.Dy() {
		ref = resize.Resize(uint(lb.Dx()), uint(lb.Dy()), reference, resize.Lanczos3)
	}
	return pixelSimilarityPct(live, ref)
}
