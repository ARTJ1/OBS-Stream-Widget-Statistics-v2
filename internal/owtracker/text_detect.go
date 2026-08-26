package owtracker

import "github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"

const (
	textMinConfidence = 0.58
	textMinMargin     = 0.20
)

func textAccepts(mr ml.Result) (string, bool) {
	if mr.Label == ml.LabelNone {
		return "", false
	}
	label := string(mr.Label)
	switch mr.Label {
	case ml.LabelWin:
		if mr.WinProb < textMinConfidence || mr.WinProb < mr.LossProb+textMinMargin {
			return label, false
		}
	case ml.LabelLoss:
		if mr.LossProb < textMinConfidence || mr.LossProb < mr.WinProb+textMinMargin {
			return label, false
		}
	}
	return label, true
}

func applyEndScreenGate(endActive bool, source string) bool {
	return endActive || source == "preview" || source == "upload"
}
