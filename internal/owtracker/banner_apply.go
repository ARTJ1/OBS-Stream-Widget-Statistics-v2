package owtracker

// instantApply is disabled — always require 2 agreeing frames to prevent false scores.
func instantApply(probe ProbeResult) bool {
	return false
}

func instantOCRApply(probe ProbeResult) bool {
	return false
}

func normalizeTriggerProbe(probe ProbeResult) ProbeResult {
	return probe
}
