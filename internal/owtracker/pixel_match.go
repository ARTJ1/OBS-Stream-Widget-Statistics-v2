package owtracker

import (
	"image"
	"math"
)

func pixelSimilarityPct(live, reference image.Image) int {
	if live == nil || reference == nil {
		return -1
	}
	a := normalizeContrast(normalizeForHash(live))
	b := normalizeContrast(normalizeForHash(reference))
	ba := a.Bounds()
	bb := b.Bounds()
	if ba.Dx() != bb.Dx() || ba.Dy() != bb.Dy() {
		return -1
	}
	var sum, sumA, sumB float64
	n := ba.Dx() * ba.Dy()
	if n == 0 {
		return -1
	}
	for y := ba.Min.Y; y < ba.Max.Y; y++ {
		for x := ba.Min.X; x < ba.Max.X; x++ {
			va := float64(a.GrayAt(x, y).Y)
			vb := float64(b.GrayAt(x, y).Y)
			diff := va - vb
			sum += diff * diff
			sumA += va
			sumB += vb
		}
	}
	mse := sum / float64(n)
	if mse < 2 {
		return 100
	}
	meanA := sumA / float64(n)
	meanB := sumB / float64(n)
	var varA, varB, cov float64
	for y := ba.Min.Y; y < ba.Max.Y; y++ {
		for x := ba.Min.X; x < ba.Max.X; x++ {
			da := float64(a.GrayAt(x, y).Y) - meanA
			db := float64(b.GrayAt(x, y).Y) - meanB
			varA += da * da
			varB += db * db
			cov += da * db
		}
	}
	corr := 0.0
	if varA > 0 && varB > 0 {
		corr = cov / math.Sqrt(varA*varB)
		if corr < 0 {
			corr = 0
		}
	}
	mseScore := 100.0 - mse/20.0
	if mseScore < 0 {
		mseScore = 0
	}
	if mseScore > 100 {
		mseScore = 100
	}
	blended := 0.55*mseScore + 0.45*(corr*100)
	score := int(math.Round(blended))
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

const (
	pixelMinScore     = 62
	pixelWinGap       = 12
	pixelAmbiguousGap = 8
	hashRelGap        = 8
	hashRelCapBonus   = 18
)

func pickMatchScores(winDist, lossDist, winPix, lossPix, threshold int, crop image.Image) (match string, wouldTrigger, ambiguous bool, method string) {
	goldPct, defeatPct := bannerColorHint(crop)
	strongGold := colorBlocksLoss(goldPct, defeatPct)

	if winPix >= 0 && lossPix >= 0 {
		weakBoth := winPix < pixelMinScore && lossPix < pixelMinScore
		if !weakBoth {
			gap := winPix - lossPix
			if gap >= pixelWinGap && winPix >= pixelMinScore && !colorBlocksWin(goldPct, defeatPct) {
				if endScreenGate("win", goldPct, defeatPct, winPix, lossPix, winDist, lossDist, threshold) {
					return "win", true, false, "pixel"
				}
			}
			if gap <= -pixelWinGap && lossPix >= pixelMinScore && !strongGold {
				if endScreenGate("loss", goldPct, defeatPct, winPix, lossPix, winDist, lossDist, threshold) {
					return "loss", true, false, "pixel"
				}
			}
			if !strongGold && absInt(gap) < pixelAmbiguousGap {
				return "", false, true, "pixel"
			}
		}
	}

	if strongGold {
		cap := threshold + hashRelCapBonus + streamCapBonus
		if winDist >= 0 && winDist <= cap {
			return "win", true, false, "color-victory"
		}
		if winDist >= 0 && lossDist >= 0 && winDist <= lossDist+hashRelGap+streamCapBonus {
			return "win", true, false, "color-victory"
		}
	}

	if winDist >= 0 && lossDist >= 0 {
		gap := lossDist - winDist
		cap := threshold + hashRelCapBonus
		if gap >= hashRelGap && winDist <= cap && !colorBlocksWin(goldPct, defeatPct) {
			if endScreenGate("win", goldPct, defeatPct, winPix, lossPix, winDist, lossDist, threshold) {
				return "win", true, false, "hash"
			}
		}
		if gap <= -hashRelGap && lossDist <= cap && !strongGold {
			if endScreenGate("loss", goldPct, defeatPct, winPix, lossPix, winDist, lossDist, threshold) {
				return "loss", true, false, "hash"
			}
		}
	}

	match, wouldTrigger, ambiguous = pickMatch(winDist, lossDist, threshold)
	if wouldTrigger {
		if match == "loss" && strongGold {
			cap := threshold + hashRelCapBonus + streamCapBonus
			if winDist >= 0 && winDist <= cap {
				return "win", true, false, "color-victory"
			}
			if winDist >= 0 && lossDist >= 0 && winDist <= lossDist+hashRelGap+streamCapBonus {
				return "win", true, false, "color-victory"
			}
			return "", false, false, "color-victory"
		}
		if match == "win" && colorBlocksWin(goldPct, defeatPct) {
			return "", false, false, "color-defeat"
		}
		if !endScreenGate(match, goldPct, defeatPct, winPix, lossPix, winDist, lossDist, threshold) {
			return "", false, false, "gameplay"
		}
		method = "hash-strict"
	}
	return match, wouldTrigger, ambiguous, method
}

func confirmsProbeMatch(first, second ProbeResult) bool {
	if first.Match == "" {
		return false
	}
	if first.Match == "loss" && colorBlocksLoss(first.GoldPct, first.DefeatPct) {
		return false
	}
	if first.Match == "win" && colorBlocksWin(first.GoldPct, first.DefeatPct) {
		return false
	}
	if second.Match == first.Match && second.WouldTrigger {
		return true
	}
	// Victory banner stays on screen while Twitch/UI shifts between frames.
	if first.Match == "win" && first.MatchMethod == "color-victory" {
		return second.GoldPct >= 6 && !colorBlocksWin(second.GoldPct, second.DefeatPct)
	}
	if first.Match == "win" {
		if second.WinPixelPct >= 0 && second.LossPixelPct >= 0 {
			return second.WinPixelPct >= second.LossPixelPct && second.WinPixelPct >= pixelMinScore-8
		}
		return second.WinDistance >= 0 && second.LossDistance >= 0 && second.WinDistance <= second.LossDistance
	}
	if first.Match == "loss" {
		if second.WinPixelPct >= 0 && second.LossPixelPct >= 0 {
			return second.LossPixelPct >= second.WinPixelPct && second.LossPixelPct >= pixelMinScore-8
		}
		return second.LossDistance >= 0 && second.WinDistance >= 0 && second.LossDistance <= second.WinDistance
	}
	return false
}

func confirmProbeConsensus(probes []ProbeResult) (ProbeResult, bool) {
	if len(probes) < 2 {
		return ProbeResult{}, false
	}
	match := probes[0].Match
	if match == "" || !probes[0].WouldTrigger {
		return ProbeResult{}, false
	}
	agree := 0
	var last ProbeResult
	for i, p := range probes {
		if i == 0 {
			agree++
			last = p
			continue
		}
		if p.Match == match && p.WouldTrigger {
			agree++
			last = p
			continue
		}
		if confirmsProbeMatch(probes[0], p) {
			agree++
			last = p
		}
	}
	if agree >= 2 {
		confirmed := probes[0]
		confirmed.GoldPct = last.GoldPct
		confirmed.DefeatPct = last.DefeatPct
		return confirmed, true
	}
	return ProbeResult{}, false
}
