package ml

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"path/filepath"
	"sync"
)

// mlpWeights is exported from scripts/ml/train_banner.py
type mlpWeights struct {
	InputSize  int       `json:"inputSize"`
	Hidden1    int       `json:"hidden1"`
	Hidden2    int       `json:"hidden2"`
	Classes    []string  `json:"classes"`
	W1         []float32 `json:"w1"`
	B1         []float32 `json:"b1"`
	W2         []float32 `json:"w2"`
	B2         []float32 `json:"b2"`
	W3         []float32 `json:"w3"`
	B3         []float32 `json:"b3"`
}

type mlpClassifier struct {
	mu      sync.Mutex
	weights mlpWeights
	ready   bool
	note    string
}

func loadMLP(dataDir string) *mlpClassifier {
	c := &mlpClassifier{}
	paths := []string{
		filepath.Join(dataDir, "banner_mlp.json"),
		"internal/owtracker/ml/assets/models/banner_mlp.json",
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &c.weights); err != nil {
			c.note = "ml parse: " + err.Error()
			return c
		}
		if len(c.weights.Classes) != 3 {
			c.note = "ml: expected 3 classes"
			return c
		}
		c.ready = true
		c.note = "mlp loaded from " + filepath.Base(p)
		return c
	}
	if b, err := embeddedMLP(); err == nil && len(b) > 0 {
		if err := json.Unmarshal(b, &c.weights); err == nil && len(c.weights.Classes) == 3 {
			c.ready = true
			c.note = "embedded mlp"
			return c
		}
	}
	c.note = "banner_mlp.json missing"
	return c
}

func (m *mlpClassifier) Ready() bool        { return m.ready }
func (m *mlpClassifier) StatusNote() string { return m.note }

func (m *mlpClassifier) Classify(img image.Image) (Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.ready {
		return Result{Label: LabelNone}, nil
	}
	x := cropToRGB48(img)
	if x == nil {
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

func denseRelu(in, w, b []float32, outN, inN int) []float32 {
	out := make([]float32, outN)
	for o := 0; o < outN; o++ {
		var sum float32
		row := o * inN
		for i := 0; i < inN; i++ {
			sum += in[i] * w[row+i]
		}
		sum += b[o]
		if sum > 0 {
			out[o] = sum
		}
	}
	return out
}

func denseLinear(in, w, b []float32, outN, inN int) []float32 {
	out := make([]float32, outN)
	for o := 0; o < outN; o++ {
		var sum float32
		row := o * inN
		for i := 0; i < inN; i++ {
			sum += in[i] * w[row+i]
		}
		out[o] = sum + b[o]
	}
	return out
}

func softmax(x []float32) []float32 {
	maxv := x[0]
	for _, v := range x[1:] {
		if v > maxv {
			maxv = v
		}
	}
	out := make([]float32, len(x))
	var sum float64
	for i, v := range x {
		e := math.Exp(float64(v - maxv))
		out[i] = float32(e)
		sum += e
	}
	for i := range out {
		out[i] /= float32(sum)
	}
	return out
}
