package owtracker

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func templateCropPath(dataDir string, kind Outcome) string {
	return filepath.Join(dataDir, "ow_templates", string(kind)+".png")
}

func saveTemplateCrop(dataDir string, kind Outcome, img image.Image) error {
	if img == nil {
		return nil
	}
	dir := filepath.Join(dataDir, "ow_templates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := templateCropPath(dataDir, kind)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func removeTemplateCrop(dataDir string, kind Outcome) {
	_ = os.Remove(templateCropPath(dataDir, kind))
}

func swapTemplateCropFiles(dataDir string) error {
	winPath := templateCropPath(dataDir, OutcomeWin)
	lossPath := templateCropPath(dataDir, OutcomeLoss)
	tmpPath := filepath.Join(dataDir, "ow_templates", "_swap.tmp")
	hasWin := fileExists(winPath)
	hasLoss := fileExists(lossPath)
	if !hasWin && !hasLoss {
		return nil
	}
	if hasWin {
		if err := os.Rename(winPath, tmpPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if hasLoss {
		if err := os.Rename(lossPath, winPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if hasWin {
		if err := os.Rename(tmpPath, lossPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (t *Tracker) TemplateCropFile(kind Outcome) string {
	saved := effectiveConfig(normalizeSaved(loadSavedHashes(t.dataDir)))
	custom := saved.CustomWin
	if kind == OutcomeLoss {
		custom = saved.CustomLoss
	}
	if !custom {
		return ""
	}
	return templateCropPath(t.dataDir, kind)
}

func (t *Tracker) BuiltinTemplateBytes(kind Outcome) ([]byte, error) {
	return defaultTemplateBytes(kind)
}
