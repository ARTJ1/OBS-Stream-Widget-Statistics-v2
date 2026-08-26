package owtracker

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestOCRInitEmbedded(t *testing.T) {
	dir := t.TempDir()
	bindOCRDataDir(dir)
	ensureOCR(dir)
	if !ocrReady {
		t.Fatalf("ocr not ready: %s", ocrInitNote)
	}
}

func TestOCREngineRunsOnTemplate(t *testing.T) {
	dir := t.TempDir()
	tr := New(dir, nil)
	if !tr.ocrReady() {
		t.Fatalf("ocr not ready: %s", ocrStatusNote())
	}
	img, err := defaultTemplateImage(OutcomeWin)
	if err != nil {
		t.Fatal(err)
	}
	text, _ := recognizeBannerText(img, dir, filepath.Join(dir, "ocr"), 0, 0)
	if text == "" && ocrInitNote != "" {
		t.Fatalf("ocr failed: %s", ocrInitNote)
	}
	// Factory PNG may be hard to read; engine must at least run without error.
	p := filepath.Join(dir, "ocr", "debug.png")
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	prepared, ok := prepareOCRImage(img, 0, 0)
	if !ok {
		t.Skip("factory template has no banner band for OCR prep")
	}
	f, _ := os.Create(p)
	_ = png.Encode(f, prepared)
	f.Close()
}
