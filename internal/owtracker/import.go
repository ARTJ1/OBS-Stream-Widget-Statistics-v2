package owtracker

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"

	"github.com/corona10/goimagehash"
)

// TemplateAnalysis is returned after importing a user screenshot as a win/loss sample.
type TemplateAnalysis struct {
	Kind              string   `json:"kind"`
	Hash              string   `json:"hash"`
	SourceW           int      `json:"sourceW"`
	SourceH           int      `json:"sourceH"`
	CropW             int      `json:"cropW"`
	CropH             int      `json:"cropH"`
	UsedFullImage     bool     `json:"usedFullImage"`
	SelfMatchPct      int      `json:"selfMatchPct"`
	VsOtherPct        int      `json:"vsOtherPct"`
	VsOtherDistance   int      `json:"vsOtherDistance"`
	QualityPct        int      `json:"qualityPct"`
	MatchThresholdPct int      `json:"matchThresholdPct"`
	OK                bool     `json:"ok"`
	Warnings          []string `json:"warnings,omitempty"`
}

func decodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("cannot decode image (use PNG or JPG): %w", err)
	}
	return img, nil
}

// templateRegion picks the same center banner zone used at runtime.
// Full screenshots are cropped; already-cropped banners are used as-is.
func templateRegion(src image.Image) (crop image.Image, usedFull bool) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 32 || h < 16 {
		return src, true
	}
	if w >= 800 && h >= 450 {
		lx, ly, cw, ch := centerBox(w, h)
		return cropRGBA(src, b.Min.X+lx, b.Min.Y+ly, cw, ch), false
	}
	return src, true
}

func cropRGBA(src image.Image, x, y, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), src, image.Pt(x, y), draw.Src)
	return dst
}

func similarityPct(distance int) int {
	if distance < 0 {
		return 0
	}
	if distance > 64 {
		distance = 64
	}
	return int(math.Round(100 * float64(64-distance) / 64))
}

func thresholdPct() int {
	return similarityPct(hashDistanceThreshold)
}

func cropContrastScore(img image.Image) int {
	b := img.Bounds()
	n := b.Dx() * b.Dy()
	if n <= 0 {
		return 0
	}
	var sum, sumSq float64
	step := 1
	if n > 20000 {
		step = 2
	}
	count := 0
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			r, g, bl, _ := img.At(x, y).RGBA()
			y8 := float64(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8))
			sum += y8
			sumSq += y8 * y8
			count++
		}
	}
	if count == 0 {
		return 0
	}
	mean := sum / float64(count)
	variance := sumSq/float64(count) - mean*mean
	if variance < 0 {
		variance = 0
	}
	stdev := math.Sqrt(variance)
	score := int(math.Round(stdev * 2.5))
	if score > 100 {
		score = 100
	}
	return score
}

func analyzeImport(kind Outcome, crop image.Image, hashStr, otherHash string, sourceW, sourceH int, usedFull bool) TemplateAnalysis {
	a := TemplateAnalysis{
		Kind:              string(kind),
		Hash:              hashStr,
		SourceW:           sourceW,
		SourceH:           sourceH,
		CropW:             crop.Bounds().Dx(),
		CropH:             crop.Bounds().Dy(),
		UsedFullImage:     usedFull,
		SelfMatchPct:      100,
		MatchThresholdPct: thresholdPct(),
		VsOtherDistance:   -1,
		OK:                true,
	}

	quality := 100
	if sourceW < 640 || sourceH < 360 {
		quality -= 15
		a.Warnings = append(a.Warnings, "low_resolution")
	}
	contrast := cropContrastScore(crop)
	if contrast < 25 {
		quality -= 35
		a.Warnings = append(a.Warnings, "low_contrast")
	} else if contrast < 45 {
		quality -= 15
		a.Warnings = append(a.Warnings, "weak_contrast")
	}
	if a.CropW < 40 || a.CropH < 16 {
		quality -= 25
		a.Warnings = append(a.Warnings, "tiny_crop")
	}

	if otherHash != "" {
		h, err := perceptionHash(crop)
		if err == nil {
			ref, err := goimagehash.ImageHashFromString(otherHash)
			if err == nil {
				if d, err := h.Distance(ref); err == nil {
					a.VsOtherDistance = d
					a.VsOtherPct = similarityPct(d)
					if d <= hashDistanceThreshold+3 {
						quality -= 40
						a.Warnings = append(a.Warnings, "too_similar_to_other")
						a.OK = false
					} else if d <= hashDistanceThreshold+10 {
						quality -= 15
						a.Warnings = append(a.Warnings, "close_to_other")
					}
				}
			}
		}
	}

	if quality < 0 {
		quality = 0
	}
	if quality < 55 {
		a.OK = false
	}
	a.QualityPct = quality
	return a
}
