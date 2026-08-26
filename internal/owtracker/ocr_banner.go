package owtracker

import (
	"image"
	"image/color"
	"strings"
	"unicode"
)

// extractBannerTextBand finds the horizontal slice that contains the large gold/red banner word.
func extractBannerTextBand(src image.Image, goldPct, defeatPct int) (image.Image, bool) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return src, false
	}
	bandH := h / 2
	if bandH < 24 {
		bandH = 24
	}
	if bandH > h {
		bandH = h
	}
	step := bandH / 4
	if step < 2 {
		step = 2
	}

	bestY, bestScore := b.Min.Y, 0
	for y := b.Min.Y; y <= b.Max.Y-bandH; y += step {
		score := scoreBannerBand(src, y, y+bandH)
		if score > bestScore {
			bestScore = score
			bestY = y
		}
	}
	minScore := w / 10
	if minScore < 20 {
		minScore = 20
	}
	strongColor := goldPct >= 20 || defeatPct >= 20
	if bestScore < minScore {
		return src, false
	}
	if !strongColor && !bandHasWideTextRow(src, bestY, bestY+bandH) {
		return src, false
	}
	return cropRect(src, b.Min.X, bestY, w, bandH), true
}

// bandHasWideTextRow requires at least one row with a long gold/red run (banner glyphs).
func bandHasWideTextRow(img image.Image, y0, y1 int) bool {
	b := img.Bounds()
	w := b.Dx()
	need := w / 6
	if need < 16 {
		need = 16
	}
	for y := y0; y < y1 && y < b.Max.Y; y++ {
		run := 0
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x4000 {
				run = 0
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			if isBannerGoldPixel(r8, g8, b8) || isBannerRedPixel(r8, g8, b8) {
				run++
				if run >= need {
					return true
				}
			} else {
				run = 0
			}
		}
	}
	return false
}

func scoreBannerBand(img image.Image, y0, y1 int) int {
	b := img.Bounds()
	score := 0
	for y := y0; y < y1 && y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x4000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			if isBannerGoldPixel(r8, g8, b8) || isBannerRedPixel(r8, g8, b8) {
				score++
			}
		}
	}
	return score
}

func isBannerGoldPixel(r, g, b int) bool {
	return r > 140 && g > 110 && b < 140 && r > b+20
}

func isBannerRedPixel(r, g, b int) bool {
	return r > 130 && r > g+25 && b < 110 && g < 130
}

func cropRect(src image.Image, x, y, w, h int) image.Image {
	b := src.Bounds()
	rect := image.Rect(x, y, x+w, y+h).Intersect(b)
	if rect.Empty() {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	for dy := 0; dy < rect.Dy(); dy++ {
		for dx := 0; dx < rect.Dx(); dx++ {
			dst.Set(dx, dy, src.At(rect.Min.X+dx, rect.Min.Y+dy))
		}
	}
	return dst
}

// bannerBinary isolates gold/red banner glyphs as white on black for OCR.
func bannerBinary(src image.Image) *image.Gray {
	b := src.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			if a < 0x3000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			if isBannerGoldPixel(r8, g8, b8) || isBannerRedPixel(r8, g8, b8) {
				out.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
	return out
}

func bannerBinaryScore(g *image.Gray) int {
	if g == nil {
		return 0
	}
	b := g.Bounds()
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if g.GrayAt(x, y).Y > 200 {
				n++
			}
		}
	}
	return n
}

func toHighContrastGray(src image.Image) *image.Gray {
	b := src.Bounds()
	gray := image.NewGray(b)
	minV, maxV := 255, 0
	vals := make([]int, b.Dx()*b.Dy())
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := src.At(x, y).RGBA()
			v := int(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8))
			vals[i] = v
			i++
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	span := maxV - minV
	if span < 24 {
		return gray
	}
	i = 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			stretched := (vals[i] - minV) * 255 / span
			gray.SetGray(x, y, color.Gray{Y: uint8(stretched)})
			i++
		}
	}
	return gray
}

func foldLatinCyrillic(s string) string {
	repl := strings.NewReplacer(
		"A", "А", "B", "В", "C", "С", "E", "Е", "H", "Н", "K", "К", "M", "М",
		"O", "О", "P", "Р", "T", "Т", "X", "Х", "Y", "У", "I", "І",
		"a", "а", "c", "с", "e", "е", "o", "о", "p", "р", "x", "х",
	)
	return repl.Replace(s)
}

func normalizeOCRText(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

func fuzzyKeywordHit(norm string, keys []string) bool {
	if containsKeyword(norm, keys) {
		return true
	}
	folded := foldLatinCyrillic(norm)
	if folded != norm && containsKeyword(folded, keys) {
		return true
	}
	compact := strings.ReplaceAll(norm, " ", "")
	for _, key := range keys {
		if len(compact) >= len(key)-2 && strings.Contains(compact, trimKey(key)) {
			return true
		}
		if len(key) >= 5 && levenshteinContains(compact, key, 2) {
			return true
		}
		if folded != norm && len(key) >= 5 && levenshteinContains(foldLatinCyrillic(compact), key, 2) {
			return true
		}
	}
	return false
}

func trimKey(key string) string {
	runes := []rune(key)
	if len(runes) > 4 {
		return string(runes[:4])
	}
	return key
}

func levenshteinContains(haystack, needle string, maxDist int) bool {
	if len(needle) < 5 || len(haystack) < len(needle)-maxDist {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if levenshtein(haystack[i:i+len(needle)], needle) <= maxDist {
			return true
		}
	}
	return false
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3int(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func min3int(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
