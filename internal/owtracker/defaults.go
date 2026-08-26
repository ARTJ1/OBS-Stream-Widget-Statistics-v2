package owtracker

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
)

//go:embed assets/templates/win.png assets/templates/loss.png
var defaultTemplateFS embed.FS

// Built-in reference templates (calibrated OW2 victory/defeat banners + zone).
// Users can upload custom samples; ResetToDefaults restores these.
var (
	defaultWinHash  = "p:c01f6b0eb8477c63"
	defaultLossHash = "p:91a07af02f3d55c3"
	defaultZone     = Zone{
		XPct: 31.630925507900677,
		YPct: 31.981733293598552,
		WPct: 39.10835214446956,
		HPct: 20.608951302661616,
	}
)

func defaultTemplateBytes(kind Outcome) ([]byte, error) {
	name := "assets/templates/" + string(kind) + ".png"
	return defaultTemplateFS.ReadFile(name)
}

func defaultTemplateImage(kind Outcome) (image.Image, error) {
	b, err := defaultTemplateBytes(kind)
	if err != nil {
		return nil, err
	}
	return decodeImage(bytes.NewReader(b))
}

func verifyDefaultTemplates() error {
	for _, kind := range []Outcome{OutcomeWin, OutcomeLoss} {
		img, err := defaultTemplateImage(kind)
		if err != nil {
			return err
		}
		h, err := computeTemplateHash(img)
		if err != nil {
			return err
		}
		want := defaultWinHash
		if kind == OutcomeLoss {
			want = defaultLossHash
		}
		if h != want {
			return ErrDefaultHashMismatch
		}
	}
	return nil
}
