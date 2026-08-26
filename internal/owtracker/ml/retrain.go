package ml

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	retrainInput  = 48
	retrainH1     = 128
	retrainH2     = 64
	retrainEpochs = 35
	retrainBatch  = 32
	retrainLR     = 0.002
	retrainNoneN  = 200
)

// Retrain rebuilds banner_mlp.json from data/ml/train and data/ml/auto samples.
func Retrain(dataDir string) (bool, string, error) {
	X, y, nWin, nLoss := gatherSamples(dataDir)
	if nWin < 2 || nLoss < 2 {
		return false, "need win+loss samples", nil
	}
	if len(y) < 8 {
		return false, "too few samples", nil
	}

	w := trainFromScratch(X, y, retrainEpochs, retrainLR)
	out := mlpWeights{
		InputSize: retrainInput,
		Hidden1:   retrainH1,
		Hidden2:   retrainH2,
		Classes:   []string{"none", "win", "loss"},
		W1:        flatten2D(w.W1, retrainH1, retrainInput),
		B1:        w.B1,
		W2:        flatten2D(w.W2, retrainH2, retrainH1),
		B2:        w.B2,
		W3:        flatten2D(w.W3, 3, retrainH2),
		B3:        w.B3,
	}
	path := filepath.Join(dataDir, "banner_mlp.json")
	b, err := json.Marshal(out)
	if err != nil {
		return false, "", err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return false, "", err
	}
	Reload(dataDir)
	note := fmt.Sprintf("%d samples (win=%d loss=%d none=%d)", len(y), nWin, nLoss, len(y)-nWin-nLoss)
	return true, note, nil
}

func Reload(dataDir string) {
	if c := loadMLP(dataDir); c != nil && c.Ready() {
		global = c
	}
}

type trainWeights struct {
	W1, B1, W2, B2, W3, B3 []float32
}

func gatherSamples(dataDir string) ([][]float32, []int, int, int) {
	classes := []struct {
		name string
		cls  int
	}{
		{"win", 1},
		{"loss", 2},
	}
	var X [][]float32
	var y []int
	nWin, nLoss := 0, 0
	for _, c := range classes {
		for _, sub := range []string{"train", "auto"} {
			dir := filepath.Join(dataDir, "ml", sub, c.name)
			_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				ext := filepath.Ext(path)
				if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
					return nil
				}
				img, err := loadImageFile(path)
				if err != nil {
					return nil
				}
				vec := cropToRGB48(img)
				if vec == nil {
					return nil
				}
				X = append(X, vec)
				y = append(y, c.cls)
				if c.name == "win" {
					nWin++
				} else {
					nLoss++
				}
				return nil
			})
		}
	}
	for i := 0; i < retrainNoneN; i++ {
		X = append(X, syntheticNone(retrainInput))
		y = append(y, 0)
	}
	return X, y, nWin, nLoss
}

func loadImageFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func syntheticNone(size int) []float32 {
	rng := rand.New(rand.NewSource(99))
	base := 15 + rng.Intn(65)
	out := make([]float32, size*size*3)
	n := size * size
	for i := 0; i < n; i++ {
		c := float32(rng.Intn(256)) / 255
		out[i] = float32(base)/255 + c*0.1
		out[n+i] = float32(base)/255 + c*0.08
		out[2*n+i] = float32(base)/255 + c*0.06
	}
	return out
}

func trainFromScratch(X [][]float32, y []int, epochs int, lr float32) trainWeights {
	n := len(y)
	d := len(X[0])
	c := 3
	rng := rand.New(rand.NewSource(42))

	w1 := randMat(rng, retrainH1, d, 0.05)
	b1 := make([]float32, retrainH1)
	w2 := randMat(rng, retrainH2, retrainH1, 0.05)
	b2 := make([]float32, retrainH2)
	w3 := randMat(rng, c, retrainH2, 0.05)
	b3 := make([]float32, c)

	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	for epoch := 0; epoch < epochs; epoch++ {
		rng.Shuffle(n, func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })
		for start := 0; start < n; start += retrainBatch {
			end := start + retrainBatch
			if end > n {
				end = n
			}
			batch := indices[start:end]
			bs := len(batch)

			var dW1 = make([][]float32, retrainH1)
			var dB1 = make([]float32, retrainH1)
			var dW2 = make([][]float32, retrainH2)
			var dB2 = make([]float32, retrainH2)
			var dW3 = make([][]float32, c)
			var dB3 = make([]float32, c)
			for i := range dW1 {
				dW1[i] = make([]float32, d)
			}
			for i := range dW2 {
				dW2[i] = make([]float32, retrainH1)
			}
			for i := range dW3 {
				dW3[i] = make([]float32, retrainH2)
			}

			for _, idx := range batch {
				x := X[idx]
				h1 := matVecRelu(w1, b1, x)
				h2 := matVecRelu(w2, b2, h1)
				logits := matVecLinear(w3, b3, h2)
				probs := vecSoftmax(logits)

				grad := make([]float32, c)
				for k := 0; k < c; k++ {
					t := float32(0)
					if k == y[idx] {
						t = 1
					}
					grad[k] = (probs[k] - t) / float32(bs)
				}

				for k := 0; k < c; k++ {
					dB3[k] += grad[k]
					for j := 0; j < retrainH2; j++ {
						dW3[k][j] += grad[k] * h2[j]
					}
				}

				dh2 := make([]float32, retrainH2)
				for j := 0; j < retrainH2; j++ {
					var s float32
					for k := 0; k < c; k++ {
						s += grad[k] * w3[k][j]
					}
					if h2[j] > 0 {
						dh2[j] = s
					}
				}
				for j := 0; j < retrainH2; j++ {
					dB2[j] += dh2[j]
					for i := 0; i < retrainH1; i++ {
						dW2[j][i] += dh2[j] * h1[i]
					}
				}

				dh1 := make([]float32, retrainH1)
				for i := 0; i < retrainH1; i++ {
					var s float32
					for j := 0; j < retrainH2; j++ {
						s += dh2[j] * w2[j][i]
					}
					if h1[i] > 0 {
						dh1[i] = s
					}
				}
				for i := 0; i < retrainH1; i++ {
					dB1[i] += dh1[i]
					for j := 0; j < d; j++ {
						dW1[i][j] += dh1[i] * x[j]
					}
				}
			}

			for i := 0; i < retrainH1; i++ {
				b1[i] -= lr * dB1[i]
				for j := 0; j < d; j++ {
					w1[i][j] -= lr * dW1[i][j]
				}
			}
			for i := 0; i < retrainH2; i++ {
				b2[i] -= lr * dB2[i]
				for j := 0; j < retrainH1; j++ {
					w2[i][j] -= lr * dW2[i][j]
				}
			}
			for k := 0; k < c; k++ {
				b3[k] -= lr * dB3[k]
				for j := 0; j < retrainH2; j++ {
					w3[k][j] -= lr * dW3[k][j]
				}
			}
		}
		_ = epoch
	}

	return trainWeights{
		W1: flattenRows(w1), B1: b1,
		W2: flattenRows(w2), B2: b2,
		W3: flattenRows(w3), B3: b3,
	}
}

func randMat(rng *rand.Rand, rows, cols int, std float64) [][]float32 {
	m := make([][]float32, rows)
	for i := range m {
		m[i] = make([]float32, cols)
		for j := range m[i] {
			m[i][j] = float32(rng.NormFloat64() * std)
		}
	}
	return m
}

func matVecRelu(w [][]float32, b, x []float32) []float32 {
	out := make([]float32, len(w))
	for i, row := range w {
		var s float32
		for j, v := range x {
			s += row[j] * v
		}
		s += b[i]
		if s > 0 {
			out[i] = s
		}
	}
	return out
}

func matVecLinear(w [][]float32, b, x []float32) []float32 {
	out := make([]float32, len(w))
	for i, row := range w {
		var s float32
		for j, v := range x {
			s += row[j] * v
		}
		out[i] = s + b[i]
	}
	return out
}

func vecSoftmax(x []float32) []float32 {
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

func flattenRows(m [][]float32) []float32 {
	var out []float32
	for _, row := range m {
		out = append(out, row...)
	}
	return out
}

func flatten2D(flat []float32, rows, cols int) []float32 {
	if len(flat) == rows*cols {
		return flat
	}
	return flat
}
