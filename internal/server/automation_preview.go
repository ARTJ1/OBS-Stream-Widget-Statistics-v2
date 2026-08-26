package server

import (
	"image/png"
	"net/http"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
)

func (s *Server) automationPreview(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	img, err := s.Auto.CapturePreview()
	if err != nil {
		code := http.StatusBadRequest
		if err == owtracker.ErrZoneRequired {
			code = http.StatusBadRequest
		}
		writeJSON(w, code, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_ = png.Encode(w, img)
}

func (s *Server) automationPreviewStatus(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	probe, err := s.Auto.PreviewProbe()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"probe": probe})
}
