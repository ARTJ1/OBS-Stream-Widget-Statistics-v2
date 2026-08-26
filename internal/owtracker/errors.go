package owtracker

import "errors"

var (
	ErrGameNotRunning     = errors.New("overwatch_not_running")
	ErrGameNotVisible     = errors.New("overwatch_not_visible")
	ErrBlankCapture       = errors.New("capture_blank")
	ErrTemplateTooSimilar = errors.New("template_too_similar_to_other")
	ErrZoneRequired       = errors.New("zone_not_configured")
	ErrZoneInvalid        = errors.New("zone_invalid")
	ErrTemplateWrongKind  = errors.New("template_wrong_kind")
	ErrDefaultHashMismatch = errors.New("builtin_template_hash_mismatch")
)
