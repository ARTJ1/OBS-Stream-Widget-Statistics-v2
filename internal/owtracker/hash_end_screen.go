package owtracker

// isHashEndScreen reports whether the center crop resembles a win/loss end banner.
func isHashEndScreen(winDist, lossDist, winPix, lossPix, threshold int) bool {
	cap := threshold + hashRelCapBonus
	if winDist >= 0 && winDist <= cap {
		return true
	}
	if lossDist >= 0 && lossDist <= cap {
		return true
	}
	if winPix >= pixelMinScore-5 {
		return true
	}
	if lossPix >= pixelMinScore-5 {
		return true
	}
	return false
}
