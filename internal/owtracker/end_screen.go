package owtracker

import "image"

// endScreenGate blocks win/loss triggers on normal gameplay HUD (dark scenes, objective
// prompts, etc.) where hash distance happens to favor one template by a few bits.
func endScreenGate(match string, goldPct, defeatPct, winPix, lossPix, winDist, lossDist, threshold int) bool {
	switch match {
	case "win":
		return allowsWinTrigger(goldPct, defeatPct, winPix, lossPix, winDist, threshold)
	case "loss":
		return allowsLossTrigger(goldPct, defeatPct, winPix, lossPix, lossDist, threshold)
	default:
		return false
	}
}

func allowsWinTrigger(goldPct, defeatPct, winPix, lossPix, winDist, threshold int) bool {
	if colorBlocksWin(goldPct, defeatPct) {
		return false
	}
	if winPix >= pixelMinScore {
		return true
	}
	if goldPct >= 8 {
		return true
	}
	if winDist >= 0 && winDist <= threshold+4 && goldPct >= 4 {
		return true
	}
	if winDist >= 0 && winDist <= threshold && brightBannerPct(goldPct, defeatPct) {
		return true
	}
	return false
}

func allowsLossTrigger(goldPct, defeatPct, winPix, lossPix, lossDist, threshold int) bool {
	if colorBlocksLoss(goldPct, defeatPct) {
		return false
	}
	if lossPix >= pixelMinScore {
		return true
	}
	if defeatPct >= 10 {
		return true
	}
	if lossDist >= 0 && lossDist <= threshold+4 && defeatPct >= 5 {
		return true
	}
	if lossDist >= 0 && lossDist <= threshold && defeatPct >= 8 {
		return true
	}
	return false
}

func brightBannerPct(goldPct, defeatPct int) bool {
	return goldPct >= 6 || defeatPct >= 6
}

// bannerColorHint detects OW2 end-screen palette in the crop.
// Victory banners are warm gold/yellow; defeat tends toward red/maroon.
func bannerColorHint(img image.Image) (goldPct, defeatPct int) {
	if img == nil {
		return 0, 0
	}
	b := img.Bounds()
	n := 0
	var gold, defeat float64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			r8, g8, b8 := float64(r>>8), float64(g>>8), float64(bl>>8)
			n++
			// Bright yellow/orange ПОБЕДА text (RU client + Twitch).
			if r8 > 165 && g8 > 115 && b8 < 130 && r8 > b8+35 {
				gold += 1.5
			} else if r8 > 140 && g8 > 100 && b8 < 140 && r8-g8 < 60 && g8 > b8 {
				gold += 1
			}
			// Gold / yellow VICTORY text (in-game + compressed Twitch).
			if r8 > 150 && g8 > 90 && b8 < 130 && r8 > b8+30 && g8 > b8 {
				gold++
			} else if r8 > 120 && g8 > 95 && b8 < 150 && r8 >= b8 && g8 > b8-15 {
				gold += 0.5
			} else if r8-float64(max8(int(g8), int(b8))) >= 28 && g8 > 70 {
				// Dark scenes / Twitch: warm text stands out by channel delta.
				gold += 0.5
			}
			// Red DEFEAT text.
			if r8 > 100 && r8 > g8+15 && b8 < 90 && g8 < 120 {
				defeat++
			} else if r8 > 85 && r8 > g8+10 && b8 < 110 && g8 < 130 {
				defeat += 0.5
			} else if r8-float64(max8(int(g8), int(b8))) >= 22 && r8 > g8+8 {
				defeat += 0.5
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	goldPct = int(gold * 100 / float64(n))
	defeatPct = int(defeat * 100 / float64(n))
	return goldPct, defeatPct
}

func colorBlocksLoss(goldPct, defeatPct int) bool {
	return goldPct >= 8 && goldPct > defeatPct+3
}

func colorBlocksWin(goldPct, defeatPct int) bool {
	return defeatPct >= 10 && defeatPct > goldPct+5
}
