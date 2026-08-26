package ml

import (
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"sync"
)

const (
	textHidden1 = 256
	textHidden2 = 64
)

var textGlobal BannerClassifier

type textMLP struct {
	mu      sync.Mutex
	weights mlpWeights
	ready   bool
	note    string
}

func InitText(dataDir string) {
	if c := loadTextMLP(dataDir); c != nil && c.Ready() {
		textGlobal = c
		return
	}
	textGlobal = &noopClassifier{}
}

func init() {
	textGlobal = &noopClassifier{}
}

func TextReady() bool { return textGlobal.Ready() }

func TextStatusNote() string { return textGlobal.StatusNote() }

// ClassifyBannerText classifies white-on-black banner word (win/loss/none), language-agnostic.
func ClassifyBannerText(crop image.Image, goldPct, defeatPct int) (Result, error) {
	g := BinaryBandFromCrop(crop, goldPct, defeatPct)
	if g == nil {
		return Result{Label: LabelNone}, nil
	}
	return textGlobal.Classify(g)
}

func loadTextMLP(dataDir string) *textMLP {
	c := &textMLP{}
	paths := []string{
		filepath.Join(dataDir, "banner_text_mlp.json"),
		"internal/owtracker/ml/assets/models/banner_text_mlp.json",
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &c.weights); err != nil {
			c.note = "text ml parse: " + err.Error()
			return c
		}
		if len(c.weights.Classes) != 3 {
			c.note = "text ml: expected 3 classes"
			return c
		}
		c.ready = true
		c.note = "text net from " + filepath.Base(p)
		return c
	}
	if b, err := embeddedTextMLP(); err == nil && len(b) > 0 {
		if err := json.Unmarshal(b, &c.weights); err == nil && len(c.weights.Classes) == 3 {
			c.ready = true
			c.note = "embedded text net"
			return c
		}
	}
	c.note = "banner_text_mlp.json missing"
	return c
}

func (m *textMLP) Ready() bool        { return m.ready }
func (m *textMLP) StatusNote() string { return m.note }

func (m *textMLP) Classify(img image.Image) (Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.ready {
		return Result{Label: LabelNone}, nil
	}
	var g *image.Gray
	switch t := img.(type) {
	case *image.Gray:
		g = t
	default:
		g = binarizeBanner(img)
	}
	x := grayToInput(g)
	if x == nil || len(x) != textInputW*textInputH {
		return Result{Label: LabelNone}, nil
	}
	h1 := denseRelu(x, m.weights.W1, m.weights.B1, m.weights.Hidden1, len(x))
	h2 := denseRelu(h1, m.weights.W2, m.weights.B2, m.weights.Hidden2, m.weights.Hidden1)
	logits := denseLinear(h2, m.weights.W3, m.weights.B3, 3, m.weights.Hidden2)
	probs := softmax(logits)
	idx := 0
	best := probs[0]
	for i := 1; i < len(probs); i++ {
		if probs[i] > best {
			best = probs[i]
			idx = i
		}
	}
	label := LabelNone
	switch m.weights.Classes[idx] {
	case "win":
		label = LabelWin
	case "loss":
		label = LabelLoss
	}
	winP, lossP, noneP := float32(0), float32(0), float32(0)
	for i, name := range m.weights.Classes {
		switch name {
		case "win":
			winP = probs[i]
		case "loss":
			lossP = probs[i]
		case "none":
			noneP = probs[i]
		}
	}
	return Result{
		Label:      label,
		Confidence: best,
		WinProb:    winP,
		LossProb:   lossP,
		NoneProb:   noneP,
	}, nil
}
