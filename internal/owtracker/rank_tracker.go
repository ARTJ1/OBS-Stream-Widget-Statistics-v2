package owtracker

import (
	"image"
	"log"
	"path/filepath"
	"time"
)

// Live auto rank (alpha): runs only during the post-match cooldown, after a
// counted win/loss, while the banner detector captures nothing.

// rankFrameWidth: OBS frame width for the reader (crops are rescaled for OCR).
const rankFrameWidth = 1280

// RankResult is the last rank the reader confirmed, shown in the admin panel.
type RankResult struct {
	RankRead
	At time.Time `json:"at"`
}

// SetRankHandler receives every confirmed rank (not called in dry run).
func (t *Tracker) SetRankHandler(fn func(RankRead)) {
	t.autoMu.Lock()
	t.onRank = fn
	t.autoMu.Unlock()
}

// SetRankAuto turns auto rank on or off (persisted in auto.json).
func (t *Tracker) SetRankAuto(on bool) {
	t.rankOn.Store(on)
	t.saveAutoConfig()
	if !on {
		t.setRankWatching(false)
	}
}

func (t *Tracker) RankAuto() bool { return t.rankOn.Load() }

func (t *Tracker) setRankWatching(on bool) {
	t.autoMu.Lock()
	t.rankWatching = on
	t.autoMu.Unlock()
}

// rankTick reads one frame for the progress screen and returns the next delay.
func (t *Tracker) rankTick(gameForeground bool, now time.Time) time.Duration {
	var r RankRead
	img, via, err := t.fullFrame(gameForeground, now)
	if err == nil {
		t.captureVia.Store(via)
		r, err = t.readRankFrame(img)
		if err != nil {
			log.Printf("owtracker: rank read: %v", err)
		}
	}
	if r.Role != "" {
		log.Printf("owtracker: rank screen role=%s tier=%s div=%d text=%q", r.Role, r.Tier, r.Division, r.RankText)
	}
	t.feedRank(now, r)
	if !t.rank.active() {
		t.setRankWatching(false)
		if t.rankByGate {
			t.rankByGate, t.rankSeen = false, RankRead{}
			t.rankGateQuiet = now.Add(rankGateQuietFor)
		}
	}
	// A fixed frame rate, not a fixed pause: players who click Continue at once
	// leave the screen up for a second or two.
	if wait := rankPoll - time.Since(now); wait > rankMinPause {
		return wait
	}
	return rankMinPause
}

// rankGateCheck runs outside the post-match watch: when the cheap gate sees what
// may be the progress screen, one full read confirms it and opens a short watch
// (a draw, a missed banner, a quick leave). full is the OBS banner frame (nil
// with screen capture).
func (t *Tracker) rankGateCheck(full image.Image, gameForeground bool, now time.Time) {
	if now.Before(t.rankGateQuiet) {
		return
	}
	var band image.Image
	if full != nil {
		band = cropRankGate(full)
	} else if gameForeground {
		band = screenRankGate()
	}
	if !rankGate(band) {
		return
	}
	// Something else keeps passing the gate (a menu, a bright map): back off.
	if t.rankGateMisses >= rankGateMaxMisses && now.Before(t.rankGateNext) {
		return
	}
	var r RankRead
	img, _, err := t.fullFrame(gameForeground, now)
	if err == nil {
		r, err = t.readRankFrame(img)
		if err != nil {
			log.Printf("owtracker: rank read: %v", err)
		}
	}
	if r.Role == "" {
		t.rankGateMisses++
		if t.rankGateMisses >= rankGateMaxMisses {
			t.rankGateNext = now.Add(rankGateBackoff)
		}
		return
	}
	t.rankGateMisses = 0
	log.Printf("owtracker: rank screen without a counted match role=%s tier=%s div=%d text=%q", r.Role, r.Tier, r.Division, r.RankText)
	t.rank.startFor(now, rankGateWatchFor)
	t.rankByGate = true
	t.rankSeen = RankRead{}
	t.setRankWatching(true)
	t.feedRank(now, r)
}

// screenRankGate captures the gate band from the game window (no OBS).
func screenRankGate() image.Image {
	x, y, w, h, ok := overwatchClientBounds()
	if !ok {
		return nil
	}
	r := rankGateRect(w, h)
	img, err := captureRegion(x+r.Min.X, y+r.Min.Y, r.Dx(), r.Dy())
	if err != nil {
		return nil
	}
	return img
}

// readRankFrame runs the reader (tests swap in a fake through t.rankReader).
func (t *Tracker) readRankFrame(img image.Image) (RankRead, error) {
	if t.rankReader != nil {
		return t.rankReader(img)
	}
	return ReadRankScreen(t.ocr, filepath.Join(t.dataDir, ".ocr"), img)
}

// feedRank passes a read to the watch and applies a confirmed rank.
func (t *Tracker) feedRank(now time.Time, r RankRead) {
	if r.Role != "" && t.rank.active() {
		t.rankSeen = r
	}
	conf, ok := t.rank.observe(now, r)
	if !ok {
		if !t.rank.active() && t.rankSeen.Role != "" && !t.rankByGate {
			// The screen was there but the rank never read: show what OCR saw (not for
			// a gate watch, which may have been something else).
			log.Printf("owtracker: rank unreadable role=%q label=%q", t.rankSeen.RoleText, t.rankSeen.RankText)
			t.autoMu.Lock()
			t.lastRankMiss = &RankResult{RankRead: t.rankSeen, At: now}
			t.autoMu.Unlock()
			t.rankSeen = RankRead{}
		}
		return
	}
	t.rankSeen = RankRead{}
	log.Printf("owtracker: RANK %s %s %d", conf.Role, conf.Tier, conf.Division)
	t.autoMu.Lock()
	t.lastRankMiss = nil
	t.lastRank = &RankResult{RankRead: conf, At: now}
	fn := t.onRank
	t.autoMu.Unlock()
	if fn != nil && !t.dryRun.Load() {
		fn(conf)
	}
}

// fullFrame is the whole game picture: from OBS (preferred) or the game window.
func (t *Tracker) fullFrame(gameForeground bool, now time.Time) (image.Image, string, error) {
	if fs := t.getFrameSource(); fs != nil && !now.Before(t.obsRetryAt) {
		img, err := fs(rankFrameWidth)
		if err == nil {
			return img, "obs", nil
		}
		t.obsRetryAt = now.Add(15 * time.Second)
	}
	if !gameForeground {
		return nil, "", ErrGameNotVisible
	}
	x, y, w, h, ok := overwatchClientBounds()
	if !ok {
		return nil, "", ErrGameNotVisible
	}
	img, err := captureRegion(x, y, w, h)
	return img, "screen", err
}
