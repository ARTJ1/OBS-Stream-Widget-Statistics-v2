//go:build !windows

package owtracker

import "errors"

func grabStrip(x, y, w, h int, fn func(bgra []byte)) error {
	return errors.New("not supported")
}
