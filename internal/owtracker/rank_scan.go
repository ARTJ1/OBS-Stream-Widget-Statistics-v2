package owtracker

import (
	"image"
	"path/filepath"
	"time"
)

// RankScanner runs the auto rank reader over recorded frames (cmd/owrank),
// with the same watch logic as the live widget.
type RankScanner struct {
	ocr *winOCR
	dir string
	w   rankWatch
}

func NewRankScanner(workDir string) *RankScanner {
	return &RankScanner{ocr: newWinOCR(workDir), dir: filepath.Join(workDir, ".ocr")}
}

func (s *RankScanner) Close() { s.ocr.Close() }

// Start opens the post-match watch (call when a win/loss is counted).
func (s *RankScanner) Start(at time.Time) { s.w.start(at) }

func (s *RankScanner) Active() bool { return s.w.active() }

// Read is one frame through the reader, without the watch.
func (s *RankScanner) Read(frame image.Image) (RankRead, error) {
	return ReadRankScreen(s.ocr, s.dir, frame)
}

// Frame reads one frame during the watch; ok is true once per match.
func (s *RankScanner) Frame(frame image.Image, at time.Time) (read, confirmed RankRead, ok bool, err error) {
	read, err = s.Read(frame)
	if err != nil {
		return read, RankRead{}, false, err
	}
	confirmed, ok = s.w.observe(at, read)
	return read, confirmed, ok, nil
}
