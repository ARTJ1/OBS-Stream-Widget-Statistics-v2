package owtracker

import (
	"image"
	"image/draw"

	"github.com/corona10/goimagehash"
)

func toGray(src image.Image) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func perceptionHash(img image.Image) (*goimagehash.ImageHash, error) {
	return goimagehash.PerceptionHash(toGray(img))
}

func hashMatches(cur *goimagehash.ImageHash, template string) (int, bool) {
	if cur == nil || template == "" {
		return -1, false
	}
	ref, err := hashFromString(template)
	if err != nil {
		return -1, false
	}
	d, err := cur.Distance(ref)
	if err != nil {
		return -1, false
	}
	return d, d <= hashDistanceThreshold
}

func hashFromString(s string) (*goimagehash.ImageHash, error) {
	return goimagehash.ImageHashFromString(s)
}

func hammingDistance(cur *goimagehash.ImageHash, template string) int {
	if cur == nil || template == "" {
		return -1
	}
	ref, err := hashFromString(template)
	if err != nil {
		return -1
	}
	d, err := cur.Distance(ref)
	if err != nil {
		return -1
	}
	return d
}
