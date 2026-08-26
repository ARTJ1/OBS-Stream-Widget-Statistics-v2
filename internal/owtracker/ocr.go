package owtracker

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sync"
)

type ocrRuntime struct {
	exePath string
	ready   bool
	note    string
}

var (
	ocrMu       sync.Mutex
	ocrCache    = map[string]ocrRuntime{}
	ocrDataDir  string
	ocrExePath  string
	ocrReady    bool
	ocrInitNote string
)

func bindOCRDataDir(dataDir string) {
	if dataDir != "" {
		ocrDataDir = dataDir
	}
}

func ensureOCR(dataDir string) {
	if dataDir == "" {
		dataDir = ocrDataDir
	}
	if dataDir == "" {
		return
	}
	ocrMu.Lock()
	defer ocrMu.Unlock()
	if r, ok := ocrCache[dataDir]; ok {
		ocrExePath, ocrReady, ocrInitNote = r.exePath, r.ready, r.note
		return
	}
	r := loadOCR(dataDir)
	ocrCache[dataDir] = r
	ocrExePath, ocrReady, ocrInitNote = r.exePath, r.ready, r.note
}

func ocrAvailable() bool {
	ensureOCR(ocrDataDir)
	return ocrReady
}

func (t *Tracker) ocrReady() bool {
	ensureOCR(t.dataDir)
	return ocrReady
}

func ocrStatusNote() string {
	ensureOCR(ocrDataDir)
	return ocrInitNote
}

func recognizeBannerText(img image.Image, dataDir, tmpDir string, goldPct, defeatPct int) (raw string, outcome string) {
	ensureOCR(dataDir)
	if img == nil || !ocrReady {
		return "", ""
	}
	variants := prepareOCRVariants(img, goldPct, defeatPct)
	if len(variants) == 0 {
		return "", ""
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", ""
	}

	bestRaw := ""
	for i, prepared := range variants {
		path := filepath.Join(tmpDir, "ocr_input.png")
		if err := writeOCRPNG(path, prepared); err != nil {
			continue
		}
		if i == 0 {
			saveOCRCrop(dataDir, prepared)
		}
		for _, psm := range []string{"8", "7", "13", "6"} {
			text, err := runTesseractPSM(path, psm)
			if err != nil {
				recordOCRError(dataDir, err)
				continue
			}
			if text == "" {
				continue
			}
			if out := parseBannerOutcome(text); out != "" {
				return text, out
			}
			if bestRaw == "" {
				bestRaw = text
			}
		}
	}
	if out := parseBannerOutcome(bestRaw); out != "" {
		return bestRaw, out
	}
	return bestRaw, ""
}

func writeOCRPNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = png.Encode(f, img)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func recordOCRError(dataDir string, err error) {
	ocrMu.Lock()
	defer ocrMu.Unlock()
	ocrInitNote = "ocr run: " + err.Error()
	if r, ok := ocrCache[dataDir]; ok {
		r.note = ocrInitNote
		ocrCache[dataDir] = r
	}
}

func saveOCRCrop(dataDir string, img image.Image) {
	if dataDir == "" || img == nil {
		return
	}
	_ = saveDebugCrop(dataDir, "ocr_last.png", img)
}

func loadOCR(dataDir string) ocrRuntime {
	exe, extracted, err := extractEmbeddedTesseract(dataDir)
	if err != nil {
		return ocrRuntime{note: "extract failed: " + err.Error()}
	}
	if !extracted {
		return ocrRuntime{note: "embedded OCR missing from exe (rebuild with scripts/build.ps1)"}
	}
	if !verifyTesseract(exe) {
		return ocrRuntime{note: "embedded tesseract invalid (rebuild exe)"}
	}
	return ocrRuntime{exePath: exe, ready: true}
}
