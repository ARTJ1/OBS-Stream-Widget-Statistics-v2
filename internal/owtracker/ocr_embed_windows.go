//go:build windows

package owtracker

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/tesseract
var embeddedTesseract embed.FS

const minTesseractExeBytes = 500_000

func extractEmbeddedTesseract(dataDir string) (exePath string, ok bool, err error) {
	if dataDir == "" {
		return "", false, errors.New("no data directory")
	}
	dest := filepath.Join(dataDir, ".ocr", "tesseract")
	exePath = filepath.Join(dest, "tesseract.exe")
	marker := filepath.Join(dest, ".embedded_v2")

	if st, statErr := os.Stat(exePath); statErr == nil && !st.IsDir() && st.Size() >= minTesseractExeBytes {
		if _, err := os.Stat(marker); err == nil {
			return exePath, true, nil
		}
	}

	if err := os.RemoveAll(dest); err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", false, err
	}

	const root = "assets/tesseract"
	walkErr := fs.WalkDir(embeddedTesseract, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			return nil
		}
		if strings.HasPrefix(filepath.Base(rel), ".") {
			return nil
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := embeddedTesseract.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.EqualFold(filepath.Base(target), "tesseract.exe") && len(data) < minTesseractExeBytes {
			return fmt.Errorf("embedded tesseract.exe too small (%d bytes)", len(data))
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		if strings.EqualFold(filepath.Base(target), "tesseract.exe") {
			_ = os.Chmod(target, 0o755)
		}
		return nil
	})
	if walkErr != nil {
		return "", false, walkErr
	}
	if err := os.WriteFile(marker, []byte("v2"), 0o644); err != nil {
		return "", false, err
	}
	st, err := os.Stat(exePath)
	if err != nil || st.IsDir() || st.Size() < minTesseractExeBytes {
		return "", false, fmt.Errorf("tesseract.exe missing after extract")
	}
	return exePath, true, nil
}
