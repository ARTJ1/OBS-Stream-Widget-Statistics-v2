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
	Win       Action = "win"
	WinDown   Action = "win_down"
	Loss      Action = "loss"
	LossDown  Action = "loss_down"
	RankUp    Action = "rank_up"
	RankDown  Action = "rank_down"
	Reset     Action = "reset"
	ModeNext  Action = "mode_next"
	RoleNext  Action = "role_next"
	Quit      Action = "quit"
)

// Handler runs for each dropped action.
type Handler func(Action)

// emptyGrace is how long an empty drop may stay before it is treated as junk.
// The Lua script renames a finished temp file into place, so an empty .wreq is
// only seen from older scripts that write in place; give them time to finish.
const emptyGrace = 2 * time.Second

// Start polls dataDir for hk_*.wreq files written by the OBS script.
// Extension is .wreq (not .cmd) so Windows never treats drops as shell scripts.
//
// Drops older than since (the process start) are discarded unapplied: they were
// written while no server was running — e.g. a "quit" from closing OBS after the
// server had already exited — and replaying them would shut the fresh server down
// or apply stale W/L presses in a burst.
func Start(ctx context.Context, dataDir string, since time.Time, handle Handler) {
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
				drain(dataDir, since, time.Now(), handle)
			}
		}
	}()
}

// drain applies every pending drop in name order (the script zero-pads time and
// sequence, so name order is press order).
func drain(dir string, since, now time.Time, handle Handler) {
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
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(since) {
			if os.Remove(path) == nil {
				log.Printf("hotkeydrop: discarded stale %s (written before start)", name)
			}
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue // still locked by the writer or antivirus; retry next tick
		}
		text := strings.TrimSpace(string(b))
		if text == "" && now.Sub(info.ModTime()) < emptyGrace {
			continue // writer may not have flushed yet
		}
		// Claim by deleting: if the file cannot be removed it is not applied, so a
		// locked drop is retried later instead of firing on every tick.
		if err := os.Remove(path); err != nil {
			continue
		}
		action := parse(text)
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
	case "win_down", "api_win_down":
		return WinDown
	case "loss", "api_loss":
		return Loss
	case "loss_down", "api_loss_down":
		return LossDown
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
