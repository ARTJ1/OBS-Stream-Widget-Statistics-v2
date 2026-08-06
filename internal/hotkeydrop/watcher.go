package hotkeydrop

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Action is a queued command from the OBS Lua script (file drop, no process spawn).
type Action string

const (
	Win      Action = "win"
	Loss     Action = "loss"
	RankUp   Action = "rank_up"
	RankDown Action = "rank_down"
	Reset    Action = "reset"
	ModeNext Action = "mode_next"
	RoleNext Action = "role_next"
	Quit     Action = "quit"
)

// Handler runs for each dropped action.
type Handler func(Action)

// Start polls dataDir for hk_*.wreq files written by the OBS script.
// Extension is .wreq (not .cmd) so Windows never treats drops as shell scripts.
func Start(ctx context.Context, dataDir string, handle Handler) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Printf("hotkeydrop: mkdir: %v", err)
		return
	}
	go func() {
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				drain(dataDir, handle)
			}
		}
	}()
}

func drain(dir string, handle Handler) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if !strings.HasPrefix(lower, "hk_") {
			continue
		}
		if !strings.HasSuffix(lower, ".wreq") && !strings.HasSuffix(lower, ".cmd") {
			continue
		}
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		_ = os.Remove(path)
		if err != nil {
			continue
		}
		action := parse(strings.TrimSpace(string(b)))
		if action == "" || handle == nil {
			continue
		}
		handle(action)
	}
}

func parse(s string) Action {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "/api/")
	s = strings.ReplaceAll(s, "/", "_")
	// rank_up?role=tank  /  rank/up?role=tank (after / → _)
	if i := strings.IndexByte(s, '?'); i >= 0 {
		base, q := s[:i], s[i+1:]
		role := ""
		for _, part := range strings.Split(q, "&") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 && kv[0] == "role" {
				role = strings.TrimSpace(kv[1])
			}
		}
		if role != "" {
			switch base {
			case "rank_up", "rankup", "api_rank_up":
				return Action("rank_up_" + role)
			case "rank_down", "rankdown", "api_rank_down":
				return Action("rank_down_" + role)
			}
		}
		s = base
	}
	switch s {
	case "win", "api_win":
		return Win
	case "loss", "api_loss":
		return Loss
	case "rank_up", "rankup", "api_rank_up":
		return RankUp
	case "rank_down", "rankdown", "api_rank_down":
		return RankDown
	case "reset", "api_reset":
		return Reset
	case "mode_next", "mode", "api_mode_next":
		return ModeNext
	case "role_next", "role", "api_role_next":
		return RoleNext
	case "quit", "exit", "shutdown", "stop":
		return Quit
	default:
		if strings.HasPrefix(s, "rank_up_") || strings.HasPrefix(s, "rank_down_") {
			return Action(s)
		}
		return ""
	}
}
