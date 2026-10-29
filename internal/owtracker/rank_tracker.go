package owtracker

import (
	"image"
	"log"
	"path/filepath"
	"time"
)

// Live auto rank (alpha): runs after a counted win/loss, during the post-match
// cooldown and for as long as the progress screen stays up after it, while the
// banner detector captures nothing.

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
	}
	// A fixed frame rate, not a fixed pause: players who click Continue at once
	// leave the screen up for a second or two.
	if wait := rankPoll - time.Since(now); wait > rankMinPause {
		return wait
	}
	return rankMinPause
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
		if !t.rank.active() && t.rankSeen.Role != "" {
			// The screen was there but the rank never read: show what OCR saw.
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
