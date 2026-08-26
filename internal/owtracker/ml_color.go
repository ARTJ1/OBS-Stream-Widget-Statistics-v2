package owtracker

import "github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"

const mlMinMargin = 0.18

// mlColorAllows checks banner palette matches the predicted outcome (gold=win, red=loss).
func mlColorAllows(label string, gold, defeat int) bool {
	return applyColorStrong(label, gold, defeat)
}

func mlMarginAllows(mr ml.Result) bool {
	switch mr.Label {
	case ml.LabelWin:
		return mr.WinProb >= mr.LossProb+mlMinMargin && mr.WinProb >= mr.NoneProb
	case ml.LabelLoss:
		return mr.LossProb >= mr.WinProb+mlMinMargin && mr.LossProb >= mr.NoneProb
	default:
		return false
	}
}

// mlHashVeto rejects ML when pHash/color strongly disagree with the label.
func mlHashVeto(r ProbeResult, label string) bool {
	if label == "loss" {
		if colorBlocksLoss(r.GoldPct, r.DefeatPct) {
			return true
		}
		if r.WinSimilarity > r.LossSimilarity+12 && r.GoldPct > r.DefeatPct+6 {
			return true
		}
	}
	if label == "win" {
		if colorBlocksWin(r.GoldPct, r.DefeatPct) {
			return true
		}
		if r.LossSimilarity > r.WinSimilarity+12 && r.DefeatPct > r.GoldPct+6 {
			return true
		}
	}
	return false
}

func mlAccepts(mr ml.Result, r ProbeResult) (string, bool) {
	if mr.Label == ml.LabelNone {
		return "", false
	}
	label := string(mr.Label)
	if mlHashVeto(r, label) {
		return label, false
	}
	if !mlColorAllows(label, r.GoldPct, r.DefeatPct) {
		return label, false
	}
	if !mlMarginAllows(mr) {
		return label, false
	}
	return label, true
}
