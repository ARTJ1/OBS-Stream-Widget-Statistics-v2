package owtracker

import (
	"image"
	"os"
	"path/filepath"
	"time"
)

// Scanner runs the production detection steps on frames from any source (cmd/owscan
// feeds recorded streams). Same stages, thresholds and once-per-match flow as the
// live tracker, but time comes from the caller (video timestamps).
type Scanner struct {
	ocr     *winOCR
	workDir string
	flow    autoFlow
	zone    Zone
}

func NewScanner(workDir string) *Scanner {
	return &Scanner{ocr: newWinOCR(workDir), workDir: workDir, zone: defaultZone}
}

func (s *Scanner) Close() { s.ocr.Close() }

// Probe is stage 1 on a small full-game frame (like OBS's 320 px frame).
func (s *Scanner) Probe(frame image.Image) (hit bool, goldPct, redPct int) {
	goldPct, redPct = stripColorPct(stripFromImage(cropBannerZone(frame, s.zone)))
	return stripLooksLikeBanner(goldPct, redPct), goldPct, redPct
}

// Read is stage 2 on a large full-game frame; it feeds the once-per-match flow.
func (s *Scanner) Read(frame image.Image, at time.Time) (BannerRead, Outcome, bool, error) {
	r, err := readBanner(s.ocr, s.workDir, cropBannerZone(frame, s.zone))
	if err != nil {
		return r, "", false, err
	}
	o, counted := s.flow.observe(at, r)
	return r, o, counted, nil
}

// Idle tells the flow that a frame had no banner (lets it re-arm).
func (s *Scanner) Idle(at time.Time) { s.flow.observe(at, BannerRead{}) }

// Cooling reports whether the flow is in the post-count pause (no capture needed).
func (s *Scanner) Cooling(at time.Time) bool {
	return !s.flow.armed() && at.Sub(s.flow.lastCount) < flowMinMatchGap
}

// ProbeAt is Probe with a custom stage-1 threshold (ground-truth passes use a low one).
func (s *Scanner) ProbeAt(frame image.Image, minPct int) (hit bool, goldPct, redPct int) {
	goldPct, redPct = stripColorPct(stripFromImage(cropBannerZone(frame, s.zone)))
	return goldPct >= minPct || redPct >= minPct, goldPct, redPct
}

// ReadOnly runs stage 2 without the once-per-match flow (ground truth).
func (s *Scanner) ReadOnly(frame image.Image) (BannerRead, error) {
	return readBanner(s.ocr, s.workDir, cropBannerZone(frame, s.zone))
}

// TextAt returns OCR text (all languages joined) of a frame region given in
// fractions of the frame (x, y, w, h), upscaled 2x for small UI text.
func (s *Scanner) TextAt(frame image.Image, x, y, w, h float64) (string, error) {
	b := frame.Bounds()
	r := image.Rect(b.Min.X+int(x*float64(b.Dx())), b.Min.Y+int(y*float64(b.Dy())),
		b.Min.X+int((x+w)*float64(b.Dx())), b.Min.Y+int((y+h)*float64(b.Dy())))
	sub, ok := frame.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return "", nil
	}
	crop := sub.SubImage(r)
	cb := crop.Bounds()
	up := image.NewRGBA(image.Rect(0, 0, cb.Dx()*2, cb.Dy()*2))
	for yy := 0; yy < cb.Dy()*2; yy++ {
		for xx := 0; xx < cb.Dx()*2; xx++ {
			up.Set(xx, yy, crop.At(cb.Min.X+xx/2, cb.Min.Y+yy/2))
		}
	}
	p := filepath.Join(s.workDir, "text_at.png")
	if err := writePNG(p, up); err != nil {
		return "", err
	}
	defer os.Remove(p)
	texts, err := s.ocr.Recognize(p)
	if err != nil {
		return "", err
	}
	out := ""
	for _, t := range texts {
		out += t + " | "
	}
	return out, nil
}

// CropBannerZone cuts the default banner zone out of a full game frame (tools).
func CropBannerZone(frame image.Image) image.Image { return cropBannerZone(frame, defaultZone) }

// CNNScanner runs the production auto win/loss path (banner CNN + vote + flow)
// on full game frames with caller-supplied times.
type CNNScanner struct {
	cnn  *BannerCNN
	vote cnnVote
	flow autoFlow
}

// NewCNNScanner uses modelPath, or the model embedded in the exe when empty.
func NewCNNScanner(modelPath string) (*CNNScanner, error) {
	var m *BannerCNN
	var err error
	if modelPath != "" {
		m, err = LoadBannerCNN(modelPath)
	} else {
		m, err = loadEmbeddedBannerCNN()
	}
	if err != nil {
		return nil, err
	}
	return &CNNScanner{cnn: m}, nil
}

// Cooling reports the post-count pause (the live widget captures nothing then).
func (s *CNNScanner) Cooling(at time.Time) bool {
	return !s.flow.armed() && at.Sub(s.flow.lastCount) < flowMinMatchGap
}

// Frame classifies one full game frame; counted is true once per match.
func (s *CNNScanner) Frame(frame image.Image, at time.Time) (r BannerRead, outcome Outcome, counted bool) {
	_, win, loss := s.cnn.Classify(cropBannerZone(frame, defaultZone))
	r = s.vote.add(at, win, loss)
	outcome, counted = s.flow.observe(at, r)
	return r, outcome, counted
}
