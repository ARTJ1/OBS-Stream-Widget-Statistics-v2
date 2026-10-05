package owtracker

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// BannerRead is one OCR attempt on a capture of the banner zone.
type BannerRead struct {
	Likely  bool    `json:"likely"`            // pre-check passed, OCR was run
	Color   Outcome `json:"color,omitempty"`   // gold → win, red → loss
	BandPct int     `json:"bandPct"`           // text band height, % of crop
	FillPct int     `json:"fillPct"`           // glyph density inside the band
	Outcome Outcome `json:"outcome,omitempty"` // accepted result (win/loss/draw)
	Word    string  `json:"word,omitempty"`    // banner word it matched
	Text    string  `json:"text,omitempty"`    // raw OCR text that matched (or last seen)
	Lang    string  `json:"lang,omitempty"`
	Variant int     `json:"variant"`
	Note    string  `json:"note,omitempty"`
	WinP    float32 `json:"winP,omitempty"`  // banner CNN probabilities
	LossP   float32 `json:"lossP,omitempty"`
}

// readBanner runs the pre-check and, if it passes, Windows OCR on the variants.
func readBanner(ocr *winOCR, workDir string, img image.Image) (BannerRead, error) {
	m := buildBannerMask(img)
	r := BannerRead{Color: m.colorOutcome(), Variant: -1}
	if m.h > 0 {
		r.BandPct = m.bandH * 100 / m.h
	}
	r.FillPct = int(m.bandFill * 100)
	if !m.likelyBanner() {
		return r, nil
	}
	r.Likely = true
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return r, err
	}
	path := filepath.Join(workDir, "banner_ocr.png")
	defer os.Remove(path) // OCR input only; nothing is kept on disk
	for i, v := range bannerVariants {
		g := renderBannerVariant(m, v)
		if g == nil {
			continue
		}
		if err := writePNG(path, g); err != nil {
			return r, err
		}
		texts, err := ocr.Recognize(path)
		if err != nil {
			return r, err
		}
		for lang, text := range texts {
			outcome, word := matchBannerWord(text)
			if outcome == "" {
				if text != "" {
					r.Text, r.Lang = text, lang
				}
				continue
			}
			// The word must agree with the banner color (gold victory / red defeat).
			if outcome != OutcomeDraw && outcome != r.Color {
				r.Note = "word/color mismatch: " + word
				continue
			}
			r.Outcome, r.Word, r.Text, r.Lang, r.Variant = outcome, word, text, lang, i
			return r, nil
		}
	}
	return r, nil
}

func writePNG(path string, img image.Image) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
