package owtracker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const debugLogMaxEntries = 300

// ProbeResult is one classification attempt (live screen, upload, or runtime poll).
type ProbeResult struct {
	At             time.Time `json:"at"`
	Source         string    `json:"source"` // live, upload, runtime, live:screen, live:window
	WindowTitle    string    `json:"windowTitle,omitempty"`
	CropW          int       `json:"cropW"`
	CropH          int       `json:"cropH"`
	WinDistance    int       `json:"winDistance"`
	LossDistance   int       `json:"lossDistance"`
	WinSimilarity  int       `json:"winSimilarityPct"`
	LossSimilarity int       `json:"lossSimilarityPct"`
	WinPixelPct    int       `json:"winPixelPct,omitempty"`
	LossPixelPct   int       `json:"lossPixelPct,omitempty"`
	MatchMethod    string    `json:"matchMethod,omitempty"`
	OcrText        string    `json:"ocrText,omitempty"`
	EndScreenActive bool     `json:"endScreenActive,omitempty"`
	GoldPct        int       `json:"goldPct,omitempty"`
	DefeatPct      int       `json:"defeatPct,omitempty"`
	AvgLuminance   int       `json:"avgLuminance,omitempty"`
	BlankCrop      bool      `json:"blankCrop,omitempty"`
	HashEndScreen  bool      `json:"hashEndScreen,omitempty"`
	MLConfidence   int       `json:"mlConfidencePct,omitempty"`
	MLWinPct       int       `json:"mlWinPct,omitempty"`
	MLLossPct      int       `json:"mlLossPct,omitempty"`
	MLNonePct      int       `json:"mlNonePct,omitempty"`
	Threshold      int       `json:"threshold"`
	Match          string    `json:"match,omitempty"`
	WouldTrigger   bool      `json:"wouldTrigger"`
	Applied        bool      `json:"applied,omitempty"`
	RejectedReason string    `json:"rejectedReason,omitempty"`
	Ambiguous      bool      `json:"ambiguous,omitempty"`
	CropSaved      string    `json:"cropSaved,omitempty"`
	HUDDetected    bool      `json:"hudDetected,omitempty"`
	State          string    `json:"state,omitempty"`
	TestMode       bool      `json:"testMode"`
	DryRun         bool      `json:"dryRun"`
	Notes          string    `json:"notes,omitempty"`
}

type DebugStatus struct {
	TestMode       bool         `json:"testMode"`
	DebugLog       bool         `json:"debugLog"`
	DryRun         bool         `json:"dryRun"`
	CaptureSource  string       `json:"captureSource"`
	MatchThreshold int          `json:"matchThreshold"`
	OcrReady       bool         `json:"ocrReady"`
	MLReady        bool         `json:"mlReady"`
	MLNote         string       `json:"mlNote,omitempty"`
	TextReady      bool         `json:"textReady"`
	TextNote       string       `json:"textNote,omitempty"`
	MLAutoWin      int          `json:"mlAutoWin,omitempty"`
	MLAutoLoss     int          `json:"mlAutoLoss,omitempty"`
	LogPath        string       `json:"logPath"`
	EntryCount     int          `json:"entryCount"`
	LastProbe      *ProbeResult `json:"lastProbe,omitempty"`
	LastApplied    *ProbeResult `json:"lastApplied,omitempty"`
}

type debugLog struct {
	mu      sync.Mutex
	dataDir string
	entries []ProbeResult
	path    string
}

func newDebugLog(dataDir string) *debugLog {
	return &debugLog{
		dataDir: dataDir,
		path:    filepath.Join(dataDir, "ow_autotest.jsonl"),
	}
}

func (d *debugLog) append(entry ProbeResult) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if entry.At.IsZero() {
		entry.At = time.Now()
	}
	d.entries = append(d.entries, entry)
	if len(d.entries) > debugLogMaxEntries {
		d.entries = d.entries[len(d.entries)-debugLogMaxEntries:]
	}

	if err := os.MkdirAll(d.dataDir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(d.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func (d *debugLog) recent(limit int) []ProbeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	if limit <= 0 || limit > len(d.entries) {
		limit = len(d.entries)
	}
	if limit == 0 {
		return nil
	}
	start := len(d.entries) - limit
	out := make([]ProbeResult, limit)
	copy(out, d.entries[start:])
	return out
}

func (d *debugLog) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}

func (d *debugLog) last() *ProbeResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.entries) == 0 {
		return nil
	}
	e := d.entries[len(d.entries)-1]
	return &e
}

func (d *debugLog) clear() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = nil
	if err := os.Remove(d.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
