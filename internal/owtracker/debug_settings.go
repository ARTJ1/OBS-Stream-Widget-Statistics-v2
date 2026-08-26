package owtracker

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type debugSettings struct {
	TestMode      bool   `json:"testMode"`
	DebugLog      bool   `json:"debugLog"`
	DryRun        bool   `json:"dryRun"`
	CaptureSource string `json:"captureSource,omitempty"`
}

func debugSettingsPath(dir string) string {
	return filepath.Join(dir, "ow_debug_settings.json")
}

func defaultDebugSettings() debugSettings {
	return debugSettings{
		TestMode:      false,
		DebugLog:      false,
		DryRun:        false,
		CaptureSource: CaptureWindow,
	}
}

func loadDebugSettings(dir string) debugSettings {
	out := defaultDebugSettings()
	b, err := os.ReadFile(debugSettingsPath(dir))
	if err != nil {
		return out
	}
	_ = json.Unmarshal(b, &out)
	if out.CaptureSource == "" {
		out.CaptureSource = CaptureWindow
	}
	return out
}

func saveDebugSettings(dir string, s debugSettings) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(debugSettingsPath(dir), b, 0o644)
}

func (t *Tracker) currentDebugSettings() debugSettings {
	return debugSettings{
		TestMode:      t.testMode.Load(),
		DebugLog:      t.debugLogOn.Load(),
		DryRun:        t.dryRun.Load(),
		CaptureSource: t.getCaptureSource(),
	}
}

func (t *Tracker) persistDebugSettings() {
	_ = saveDebugSettings(t.dataDir, t.currentDebugSettings())
}

func (t *Tracker) applyDebugSettings(s debugSettings) {
	t.testMode.Store(s.TestMode)
	t.debugLogOn.Store(s.DebugLog)
	t.dryRun.Store(s.DryRun)
	t.SetCaptureSource(s.CaptureSource)
}
