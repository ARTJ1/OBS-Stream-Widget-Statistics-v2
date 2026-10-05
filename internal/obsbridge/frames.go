package obsbridge

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"time"

	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sources"
)

// Frames for auto win/loss come from OBS itself: OBS already captures the game for
// the stream, so we only ask it for a small scaled copy (scaled on the GPU by OBS).
// Nothing is written to disk.

var ErrNoGameSource = errors.New("no game capture source showing in OBS")

const gameSourceRecheck = 30 * time.Second

// gameSourceRank: higher is better; 0 = not a game picture.
func gameSourceRank(kind string, settings map[string]any) int {
	window := strings.ToLower(fmt.Sprint(settings["window"]))
	isOW := strings.Contains(window, "overwatch")
	switch {
	case strings.HasPrefix(kind, "game_capture") && isOW:
		return 5
	case strings.HasPrefix(kind, "window_capture") && isOW:
		return 4
	case strings.HasPrefix(kind, "game_capture"):
		return 3 // e.g. "capture any fullscreen application"
	case strings.HasPrefix(kind, "monitor_capture"):
		return 2
	}
	return 0
}

// GameSource returns the name of the OBS input that shows the game in Program.
// The choice is cached and re-checked periodically.
func (b *Bridge) GameSource() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.gameSource != "" && time.Since(b.gameSourceAt) < gameSourceRecheck {
		return b.gameSource, nil
	}
	client, _, err := b.ensureClient()
	if err != nil {
		return "", err
	}
	list, err := client.Inputs.GetInputList(inputs.NewGetInputListParams())
	if err != nil {
		b.dropClient(err)
		return "", err
	}
	best, bestRank := "", 0
	for _, in := range list.Inputs {
		settings := map[string]any{}
		if resp, err := client.Inputs.GetInputSettings(inputs.NewGetInputSettingsParams().WithInputName(in.InputName)); err == nil {
			settings = resp.InputSettings
		}
		rank := gameSourceRank(in.InputKind, settings)
		if rank <= bestRank {
			continue
		}
		active, err := client.Sources.GetSourceActive(sources.NewGetSourceActiveParams().WithSourceName(in.InputName))
		if err != nil || !active.VideoActive {
			continue // not on the live (Program) scene
		}
		best, bestRank = in.InputName, rank
	}
	b.gameSource, b.gameSourceAt = best, time.Now()
	if best == "" {
		return "", ErrNoGameSource
	}
	return best, nil
}

// SourceFrame asks OBS for a frame of the source scaled to width (aspect kept);
// width <= 0 returns the source at its native size.
func (b *Bridge) SourceFrame(sourceName string, width int) (image.Image, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	client, _, err := b.ensureClient()
	if err != nil {
		return nil, err
	}
	params := sources.NewGetSourceScreenshotParams().WithSourceName(sourceName)
	if width > 0 && width <= 960 {
		// Small frames (auto win/loss runs ~4 per second): JPEG is much cheaper for
		// OBS to encode than PNG, and the classifier was trained on stream video.
		params = params.WithImageFormat("jpg").WithImageCompressionQuality(90)
	} else {
		params = params.WithImageFormat("png")
	}
	if width > 0 {
		params = params.WithImageWidth(float64(width))
	}
	resp, err := client.Sources.GetSourceScreenshot(params)
	if err != nil {
		b.gameSource = "" // source renamed/removed: find it again next time
		return nil, err
	}
	data := resp.ImageData
	if i := strings.Index(data, ","); i >= 0 && strings.HasPrefix(data, "data:") {
		data = data[i+1:]
	}
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	return img, err
}

// dropClient forgets a broken connection; caller holds b.mu.
func (b *Bridge) dropClient(err error) {
	b.lastErr = err.Error()
	if b.client != nil {
		_ = b.client.Disconnect()
		b.client = nil
	}
}
