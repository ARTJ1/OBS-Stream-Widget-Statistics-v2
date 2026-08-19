package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
)

func (s *Server) automationToggle(w http.ResponseWriter, r *http.Request) {
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if q := strings.TrimSpace(r.URL.Query().Get("enabled")); q != "" {
			s.Auto.SetEnabled(q == "1" || strings.EqualFold(q, "true") || q == "on")
		}
		writeJSON(w, http.StatusOK, s.Auto.Status())
	case http.MethodPost, http.MethodPut:
		if q := strings.TrimSpace(r.URL.Query().Get("enabled")); q != "" {
			s.Auto.SetEnabled(q == "1" || strings.EqualFold(q, "true") || q == "on")
		} else {
			s.Auto.Toggle()
		}
		writeJSON(w, http.StatusOK, s.Auto.Status())
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) automationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		writeJSON(w, http.StatusOK, owtracker.Status{State: "inactive"})
		return
	}
	writeJSON(w, http.StatusOK, s.Auto.Status())
}

func (s *Server) automationCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	kind := owtracker.Outcome(strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind"))))
	if kind != owtracker.OutcomeWin && kind != owtracker.OutcomeLoss {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind=win or kind=loss required"})
		return
	}
	del := r.Method == http.MethodDelete || r.URL.Query().Get("delete") == "1"
	if del {
		if err := s.Auto.ClearTemplate(kind); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": kind, "status": s.Auto.Status()})
		return
	}
	delay := 5
	if q := r.URL.Query().Get("delay"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n >= 0 && n <= 30 {
			delay = n
		}
	}
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}
	hash, err := s.Auto.CaptureTemplate(kind)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, owtracker.ErrGameNotRunning) || errors.Is(err, owtracker.ErrGameNotVisible) {
			code = http.StatusConflict
		}
		writeJSON(w, code, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"kind":   kind,
		"hash":   hash,
		"status": s.Auto.Status(),
	})
}
