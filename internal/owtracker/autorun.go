package owtracker

import (
	"encoding/json"
	"image"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Production auto win/loss: capture the banner zone of the Overwatch window,
// read it with Windows OCR, count once per match (autoFlow).

// Detection runs the banner CNN on the banner zone 4 times a second (~2 ms each),
// on a 480 px frame from OBS (or the game window as a fallback). Nothing is
// captured during the post-match cooldown.
const (
	pollCNN     = 250 * time.Millisecond
	obsCNNWidth = 480
)

// AutoResult is the last counted (or draw) banner, shown in the admin panel.
type AutoResult struct {
	Outcome Outcome   `json:"outcome"`
	Word    string    `json:"word,omitempty"`
	At      time.Time `json:"at"`
}

// bannerBounds maps the zone onto the 16:9 area of the game client, so
// ultrawide (letterboxed HUD) and 16:10/4:3 clients crop the same banner.
func bannerBounds(x0, y0, w, h int) (int, int, int, int) {
	const ar = 16.0 / 9.0
	if w <= 0 || h <= 0 {
		return x0, y0, w, h
	}
	if float64(w)/float64(h) > ar+0.01 {
		nw := int(float64(h) * ar)
		return x0 + (w-nw)/2, y0, nw, h
	}
	if float64(w)/float64(h) < ar-0.01 {
		nh := int(float64(w) / ar)
		return x0, y0 + (h-nh)/2, w, nh
	}
	return x0, y0, w, h
}

// bannerScreenRect is the banner zone in screen pixels for the Overwatch window.
func bannerScreenRect(zone Zone) (x, y, w, h int, err error) {
	x0, y0, bw, bh, ok := overwatchClientBounds()
	if !ok {
		return 0, 0, 0, 0, ErrGameNotVisible
	}
	x0, y0, bw, bh = bannerBounds(x0, y0, bw, bh)
	lx, ly, w, h := zone.PixelBox(bw, bh)
	return x0 + lx, y0 + ly, w, h, nil
}

// FrameSource returns the game picture scaled to the given width (from OBS).
type FrameSource func(width int) (image.Image, error)

// SetFrameSource makes OBS the primary frame source (nil = screen capture only).
func (t *Tracker) SetFrameSource(fs FrameSource) {
	t.autoMu.Lock()
	t.frameSrc = fs
	t.autoMu.Unlock()
}

func (t *Tracker) getFrameSource() FrameSource {
	t.autoMu.Lock()
	defer t.autoMu.Unlock()
	return t.frameSrc
}

// cropBannerZone cuts the banner zone out of a full game frame.
func cropBannerZone(img image.Image, zone Zone) image.Image {
	b := img.Bounds()
	x, y, w, h := bannerBounds(0, 0, b.Dx(), b.Dy())
	lx, ly, cw, ch := zone.PixelBox(w, h)
	r := image.Rect(b.Min.X+x+lx, b.Min.Y+y+ly, b.Min.X+x+lx+cw, b.Min.Y+y+ly+ch)
	if sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		return sub.SubImage(r)
	}
	return img
}

// autoTick runs one production detection step and returns the next poll delay.
// gameForeground tells whether the screen-capture fallback may be used.
func (t *Tracker) autoTick(gameForeground bool) time.Duration {
	now := time.Now()
	t.autoMu.Lock()
	cooling := !t.flow.armed() && now.Sub(t.flow.lastCount) < flowMinMatchGap
	t.autoMu.Unlock()
	if cooling {
		// Just counted a match: no new one can end this soon, capture nothing.
		t.state.Store(StateMenu)
		return 5 * time.Second
	}
	if t.cnn == nil {
		t.state.Store(StateInactive)
		return pollInactive
	}
	zone := t.getZone()
	if !zone.Valid() {
		zone = defaultZone
	}
	img, via, err := t.zoneFrame(zone, gameForeground, now)
	if err != nil {
		t.captureVia.Store("")
		t.state.Store(StateInactive)
		return pollInactive
	}
	t.captureVia.Store(via)
	_, win, loss := t.cnn.Classify(img)
	t.handleRead(t.vote.add(now, win, loss))
	return pollCNN
}

// zoneFrame returns the banner zone: from OBS (preferred) or the game window.
func (t *Tracker) zoneFrame(zone Zone, gameForeground bool, now time.Time) (image.Image, string, error) {
	if fs := t.getFrameSource(); fs != nil && !now.Before(t.obsRetryAt) {
		img, err := fs(obsCNNWidth)
		if err == nil {
			return cropBannerZone(img, zone), "obs", nil
		}
		// OBS closed, not connected, or no game source on the live scene:
		// use the screen fallback and try OBS again a bit later.
		t.obsRetryAt = now.Add(15 * time.Second)
	}
	if !gameForeground {
		return nil, "", ErrGameNotVisible
	}
	x, y, w, h, err := bannerScreenRect(zone)
	if err != nil {
		return nil, "", err
	}
	img, err := captureRegion(x, y, w, h)
	return img, "screen", err
}

// handleRead feeds the once-per-match flow and applies a counted result.
func (t *Tracker) handleRead(read BannerRead) {
	now := time.Now()
	t.autoMu.Lock()
	t.lastRead = read
	t.lastReadAt = now
	outcome, counted := t.flow.observe(now, read)
	armed := t.flow.armed()
	t.autoMu.Unlock()

	if armed {
		t.state.Store(StateMatch)
	} else {
		t.state.Store(StateMenu)
	}
	if read.Likely {
		log.Printf("owtracker: banner win=%.2f loss=%.2f outcome=%q", read.WinP, read.LossP, read.Outcome)
	}
	if counted {
		res := AutoResult{Outcome: outcome, At: now}
		t.autoMu.Lock()
		t.lastResult = &res
		t.autoMu.Unlock()
		log.Printf("owtracker: COUNTED %s", outcome)
		if t.onResult != nil && !t.dryRun.Load() {
			t.onResult(outcome)
		}
	}
}

// --- persisted on/off switch (streamers turn it on once) ---

func autoConfigPath(dataDir string) string { return filepath.Join(dataDir, "auto.json") }

func loadAutoEnabled(dataDir string) bool {
	b, err := os.ReadFile(autoConfigPath(dataDir))
	if err != nil {
		return false
	}
	var v struct {
		Enabled bool `json:"enabled"`
	}
	return json.Unmarshal(b, &v) == nil && v.Enabled
}

func saveAutoEnabled(dataDir string, on bool) {
	b, _ := json.Marshal(map[string]bool{"enabled": on})
	if err := os.WriteFile(autoConfigPath(dataDir), b, 0o644); err != nil {
		log.Printf("owtracker: save auto.json: %v", err)
	}
}

// ReadBannerImage runs the production banner reader on a file image: a full
// game screenshot (cropped to the banner zone) or an already cropped zone.
func (t *Tracker) ReadBannerImage(img image.Image) (BannerRead, error) {
	return ReadBannerWith(t.ocr, t.dataDir, img)
}

// ReadBannerWith is ReadBannerImage for tools (cmd/owbench) without a Tracker.
func ReadBannerWith(ocr *winOCR, dataDir string, img image.Image) (BannerRead, error) {
	b := img.Bounds()
	if b.Dx() >= 1280 && float64(b.Dx())/float64(b.Dy()) >= 1.3 {
		img = cropBannerZone(img, defaultZone)
	}
	return readBanner(ocr, filepath.Join(dataDir, ".ocr"), img)
}

// NewWindowsOCR exposes the OCR process for tools; Close it when done.
func NewWindowsOCR(dataDir string) *winOCR { return newWinOCR(dataDir) }
