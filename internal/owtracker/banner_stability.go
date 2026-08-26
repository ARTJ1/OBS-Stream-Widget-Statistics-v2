package owtracker

import (
	"image"
	"sync"
	"time"
)

const (
	stabilityWindow      = 3000 * time.Millisecond
	stabilityMinAgree    = 3
	stabilityMinAgreeML  = 2
	stabilityHashMax     = 14
)

type bannerSample struct {
	at     time.Time
	probe  ProbeResult
	hash   string
	hashOK bool
}

// bannerStability tracks recent end-screen probes and confirms when frames agree.
type bannerStability struct {
	mu      sync.Mutex
	samples []bannerSample
}

func (b *bannerStability) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.samples = nil
}

func (b *bannerStability) add(probe ProbeResult, crop image.Image) {
	if probe.Match != "win" && probe.Match != "loss" {
		return
	}
	if probe.MatchMethod != "ocr" && probe.MatchMethod != "banner" && probe.MatchMethod != "ml" && probe.MatchMethod != "text" {
		return
	}
	if !applyProbeAllowed(probe) {
		return
	}
	if probe.At.IsZero() {
		probe.At = time.Now()
	}
	h, err := computeTemplateHash(crop)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.samples = append(b.samples, bannerSample{
		at:     probe.At,
		probe:  probe,
		hash:   h,
		hashOK: err == nil && h != "",
	})
	if len(b.samples) > 8 {
		b.samples = b.samples[len(b.samples)-8:]
	}
}

func (b *bannerStability) confirm(now time.Time) (ProbeResult, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	cutoff := now.Add(-stabilityWindow)
	var recent []bannerSample
	for _, s := range b.samples {
		if !s.at.Before(cutoff) {
			recent = append(recent, s)
		}
	}
	if len(recent) < minAgreeFor(probeMethod(recent)) {
		return ProbeResult{}, false
	}

	minAgree := minAgreeFor(probeMethod(recent))
	win, loss := 0, 0
	var lastWin, lastLoss ProbeResult
	for _, s := range recent {
		if !applyProbeAllowed(s.probe) {
			continue
		}
		switch s.probe.Match {
		case "win":
			win++
			lastWin = s.probe
		case "loss":
			loss++
			lastLoss = s.probe
		}
	}
	if win >= minAgree && win > loss {
		return lastWin, true
	}
	if loss >= minAgree && loss > win {
		return lastLoss, true
	}

	// Two consecutive similar frames with the same result.
	for i := 1; i < len(recent); i++ {
		a, c := recent[i-1], recent[i]
		if a.probe.Match == "" || a.probe.Match != c.probe.Match {
			continue
		}
		if !applyProbeAllowed(c.probe) {
			continue
		}
		if a.hashOK && c.hashOK && hashNear(a.hash, c.hash, stabilityHashMax) {
			if agreeCount(recent, c.probe.Match) >= minAgree {
				return c.probe, true
			}
		}
	}
	return ProbeResult{}, false
}

func probeMethod(recent []bannerSample) string {
	if len(recent) == 0 {
		return ""
	}
	return recent[len(recent)-1].probe.MatchMethod
}

func minAgreeFor(method string) int {
	if method == "text" || method == "ml" {
		return stabilityMinAgreeML
	}
	return stabilityMinAgree
}

func agreeCount(recent []bannerSample, match string) int {
	n := 0
	for _, s := range recent {
		if s.probe.Match == match && applyProbeAllowed(s.probe) {
			n++
		}
	}
	return n
}

func hashNear(a, b string, maxDist int) bool {
	if a == "" || b == "" {
		return false
	}
	ha, err1 := hashFromString(a)
	hb, err2 := hashFromString(b)
	if err1 != nil || err2 != nil {
		return a == b
	}
	d, err := ha.Distance(hb)
	return err == nil && d <= maxDist
}
