package owtracker

import (
	"image"
	"image/color"

	"github.com/corona10/goimagehash"
	"github.com/nfnt/resize"
)

// Fixed size for all template/live comparisons — avoids aspect-ratio skew when
// win/loss samples were uploaded at different resolutions.
const hashNormW, hashNormH = 320, 120

// Stream/compressed video: allow a higher cap only when one side is clearly closer.
const (
	streamCapBonus = 8
	streamMinGap   = 12
)

func normalizeForHash(img image.Image) image.Image {
	b := img.Bounds()
	if b.Dx() == hashNormW && b.Dy() == hashNormH {
		return img
	}
	return resize.Resize(hashNormW, hashNormH, img, resize.Bilinear)
}

// normalizeContrast stretches luminance so Twitch/OBS compression does not wash
// out the banner before hashing.
func normalizeContrast(src image.Image) *image.Gray {
	gray := toGray(src)
	b := gray.Bounds()
	minV, maxV := 255, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			v := int(gray.GrayAt(x, y).Y)
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	span := maxV - minV
	if span < 16 {
		return gray
	}
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			v := int(gray.GrayAt(x, y).Y)
			stretched := (v - minV) * 255 / span
			out.SetGray(x, y, color.Gray{Y: uint8(stretched)})
		}
	}
	return out
}

func perceptionHashLive(img image.Image) (*goimagehash.ImageHash, error) {
	return goimagehash.PerceptionHash(normalizeContrast(img))
}

func computeTemplateHash(img image.Image) (string, error) {
	norm := normalizeForHash(img)
	h, err := perceptionHashLive(norm)
	if err != nil {
		return "", err
	}
	return h.ToString(), nil
}

func bestDistance(img image.Image, template string) int {
	if template == "" {
		return -1
	}
	norm := normalizeForHash(img)
	h, err := perceptionHashLive(norm)
	if err != nil {
		return -1
	}
	return hammingDistance(h, template)
}

func matchesTemplate(img image.Image, template string, threshold int) (int, bool) {
	d := bestDistance(img, template)
	if d < 0 {
		return -1, false
	}
	return d, d <= threshold
}

func pickMatch(winDist, lossDist, threshold int) (match string, wouldTrigger, ambiguous bool) {
	if winDist < 0 && lossDist < 0 {
		return "", false, false
	}
	if winDist >= 0 && lossDist >= 0 {
		gap := absInt(winDist - lossDist)
		if gap < matchDiscrimination {
			return "", false, true
		}
		if winDist < lossDist && winDist <= threshold {
			return "win", true, false
		}
		if lossDist < winDist && lossDist <= threshold {
			return "loss", true, false
		}
		streamCap := threshold + streamCapBonus
		if winDist < lossDist && winDist <= streamCap && gap >= streamMinGap {
			return "win", true, false
		}
		if lossDist < winDist && lossDist <= streamCap && gap >= streamMinGap {
			return "loss", true, false
		}
		return "", false, false
	}
	if winDist >= 0 && winDist <= threshold {
		return "win", true, false
	}
	if lossDist >= 0 && lossDist <= threshold {
		return "loss", true, false
	}
	return "", false, false
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
