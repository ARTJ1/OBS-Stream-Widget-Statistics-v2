package owtracker

import "errors"

var (
	ErrGameNotRunning     = errors.New("overwatch_not_running")
	ErrGameNotVisible     = errors.New("overwatch_not_visible")
	ErrTemplateTooSimilar = errors.New("template_too_similar_to_other")
)
