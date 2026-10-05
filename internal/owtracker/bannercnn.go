package owtracker

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/nfnt/resize"
)

// Banner CNN: a tiny convolutional net (scripts/ml/train_banner_cnn.py) that
// classifies the banner zone as none / win / loss from the word SHAPE. It is
// trained with color, brightness and occlusion augmentation, so dim defeat
// banners and menus over the word still read. ~2M multiply-adds per frame.

const (
	cnnSrcW, cnnSrcH = 128, 40 // dataset size (cmd/owdump)
)

type BannerCNN struct {
	w, h    int
	classes []string
	arch    [][2]int
	hidden  int
	p       map[string][]float32
}

type cnnFile struct {
	W       int                  `json:"w"`
	H       int                  `json:"h"`
	Classes []string             `json:"classes"`
	Arch    [][2]int             `json:"arch"`
	Hidden  int                  `json:"fc_hidden"`
	Params  map[string][]float32 `json:"params"`
}

func LoadBannerCNN(path string) (*BannerCNN, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseBannerCNN(b)
}

func parseBannerCNN(b []byte) (*BannerCNN, error) {
	var f cnnFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if len(f.Classes) != 3 || len(f.Arch) == 0 {
		return nil, fmt.Errorf("banner cnn: bad model")
	}
	return &BannerCNN{w: f.W, h: f.H, classes: f.Classes, arch: f.Arch, hidden: f.Hidden, p: f.Params}, nil
}

// prepare: zone crop -> 128x40 (nfnt bilinear, as owdump) -> 96x32 (same bilinear
// sampling as training) -> per-image standardization, CHW.
func (m *BannerCNN) prepare(zone image.Image) []float32 {
	src := resize.Resize(cnnSrcW, cnnSrcH, zone, resize.Bilinear)
	px := make([]float32, cnnSrcW*cnnSrcH*3)
	for y := 0; y < cnnSrcH; y++ {
		for x := 0; x < cnnSrcW; x++ {
			c := color.RGBAModel.Convert(src.At(x, y)).(color.RGBA)
			i := (y*cnnSrcW + x) * 3
			px[i], px[i+1], px[i+2] = float32(c.R), float32(c.G), float32(c.B)
		}
	}
	w, h := m.w, m.h
	hwc := make([]float32, w*h*3)
	for oy := 0; oy < h; oy++ {
		sy := (float64(oy)+0.5)*cnnSrcH/float64(h) - 0.5
		y0 := clampInt(int(math.Floor(sy)), 0, cnnSrcH-1)
		y1 := clampInt(y0+1, 0, cnnSrcH-1)
		wy := float32(sy - float64(y0))
		for ox := 0; ox < w; ox++ {
			sx := (float64(ox)+0.5)*cnnSrcW/float64(w) - 0.5
			x0 := clampInt(int(math.Floor(sx)), 0, cnnSrcW-1)
			x1 := clampInt(x0+1, 0, cnnSrcW-1)
			wx := float32(sx - float64(x0))
			for c := 0; c < 3; c++ {
				a := px[(y0*cnnSrcW+x0)*3+c]*(1-wx) + px[(y0*cnnSrcW+x1)*3+c]*wx
				b := px[(y1*cnnSrcW+x0)*3+c]*(1-wx) + px[(y1*cnnSrcW+x1)*3+c]*wx
				hwc[(oy*w+ox)*3+c] = a*(1-wy) + b*wy
			}
		}
	}
	var sum, sq float64
	for _, v := range hwc {
		sum += float64(v)
	}
	mean := sum / float64(len(hwc))
	for _, v := range hwc {
		d := float64(v) - mean
		sq += d * d
	}
	std := float32(math.Sqrt(sq/float64(len(hwc)))) + 8
	chw := make([]float32, len(hwc))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			for c := 0; c < 3; c++ {
				chw[c*h*w+y*w+x] = (hwc[(y*w+x)*3+c] - float32(mean)) / std
			}
		}
	}
	return chw
}

// Classify returns probabilities for none / win / loss.
func (m *BannerCNN) Classify(zone image.Image) (none, win, loss float32) {
	x := m.prepare(zone)
	c, h, w := 3, m.h, m.w
	for i, a := range m.arch {
		co := a[1]
		wt, bias := m.p[fmt.Sprintf("c%dw", i)], m.p[fmt.Sprintf("c%db", i)]
		out := make([]float32, co*(h/2)*(w/2))
		conv := make([]float32, h*w)
		for o := 0; o < co; o++ {
			for j := range conv {
				conv[j] = bias[o]
			}
			for ci := 0; ci < c; ci++ {
				in := x[ci*h*w : (ci+1)*h*w]
				k := wt[(o*c+ci)*9 : (o*c+ci)*9+9]
				// One pass per kernel tap over whole rows (zero padding = skipped edges).
				for ky := -1; ky <= 1; ky++ {
					y0, y1 := 0, h
					if ky < 0 {
						y0 = 1
					} else if ky > 0 {
						y1 = h - 1
					}
					for kx := -1; kx <= 1; kx++ {
						kv := k[(ky+1)*3+kx+1]
						x0, x1 := 0, w
						if kx < 0 {
							x0 = 1
						} else if kx > 0 {
							x1 = w - 1
						}
						n := x1 - x0
						for y := y0; y < y1; y++ {
							dst := conv[y*w+x0 : y*w+x0+n]
							src := in[(y+ky)*w+x0+kx : (y+ky)*w+x0+kx+n]
							for j := range dst {
								dst[j] += kv * src[j]
							}
						}
					}
				}
			}
			// ReLU + 2x2 max pool
			oh, ow := h/2, w/2
			for y := 0; y < oh; y++ {
				for xx := 0; xx < ow; xx++ {
					mx := float32(0)
					for dy := 0; dy < 2; dy++ {
						for dx := 0; dx < 2; dx++ {
							if v := conv[(2*y+dy)*w+2*xx+dx]; v > mx {
								mx = v
							}
						}
					}
					out[o*oh*ow+y*ow+xx] = mx
				}
			}
		}
		x, c, h, w = out, co, h/2, w/2
	}
	hidden := make([]float32, m.hidden)
	f0w, f0b := m.p["f0w"], m.p["f0b"]
	for j := 0; j < m.hidden; j++ {
		s := f0b[j]
		for i, v := range x {
			s += v * f0w[i*m.hidden+j]
		}
		if s < 0 {
			s = 0
		}
		hidden[j] = s
	}
	f1w, f1b := m.p["f1w"], m.p["f1b"]
	var logits [3]float64
	for k := 0; k < 3; k++ {
		s := f1b[k]
		for j, v := range hidden {
			s += v * f1w[j*3+k]
		}
		logits[k] = float64(s)
	}
	mx := math.Max(logits[0], math.Max(logits[1], logits[2]))
	var e [3]float64
	var tot float64
	for k := range e {
		e[k] = math.Exp(logits[k] - mx)
		tot += e[k]
	}
	return float32(e[0] / tot), float32(e[1] / tot), float32(e[2] / tot)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
