package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/deck"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/hub"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/obsbridge"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/runtimeinfo"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/server"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/skins"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
	webassets "github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/web"
)

func TestDeckStateEndpoint(t *testing.T) {
	dir := t.TempDir()
	st, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	skinStore, err := skins.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	info := runtimeinfo.Build("127.0.0.1", 19123, "test")
	srv := server.New(st, skinStore, hub.New(), obsbridge.New(), info, webassets.FS, nil)

	if _, err := st.AddWin(); err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddWin(); err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddLoss(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/deck/state", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var d deck.State
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.Wins != 2 || d.Losses != 1 {
		t.Fatalf("wl=%d/%d", d.Wins, d.Losses)
	}
	if d.RankLabel == "" || d.RankImageURL == "" {
		t.Fatalf("rank empty: %+v", d)
	}
	wantPrefix := "http://127.0.0.1:19123/overlay/"
	if len(d.RankImageURL) < len(wantPrefix) || d.RankImageURL[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("rankImageUrl=%q", d.RankImageURL)
	}
	_ = filepath.Base(d.RankImageURL)
}
