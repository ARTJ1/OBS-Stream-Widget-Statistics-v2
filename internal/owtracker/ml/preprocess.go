package ml

import (
	"image"
	"image/color"

	"github.com/nfnt/resize"
)

const inputSize = 48

// cropToRGB48 resizes crop to 48x48 and returns normalized RGB [0,1] in CHW planar order (matches train_banner.py).
func cropToRGB48(img image.Image) []float32 {
	if img == nil {
		return nil
	}
	small := resize.Resize(inputSize, inputSize, img, resize.Lanczos3)
	n := inputSize * inputSize
	out := make([]float32, n*3)
	b := small.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.RGBAModel.Convert(small.At(x, y)).(color.RGBA)
			idx := (y-b.Min.Y)*inputSize + (x - b.Min.X)
			out[idx] = float32(c.R) / 255
			out[n+idx] = float32(c.G) / 255
			out[2*n+idx] = float32(c.B) / 255
		}
	}
	return out
}
