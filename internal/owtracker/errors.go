package owtracker

import "errors"

var (
	ErrGameNotRunning = errors.New("overwatch_not_running")
	ErrGameNotVisible = errors.New("overwatch_not_visible")
)
