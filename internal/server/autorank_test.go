package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/hub"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/obsbridge"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/runtimeinfo"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/server"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/skins"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
	webassets "github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/web"
)

type autoRankEnv struct {
	srv *server.Server
	st  *store.Store
	h   *hub.Hub
	dir string
}

func newAutoRankEnv(t *testing.T, mode string) autoRankEnv {
	t.Helper()
	dir := t.TempDir()
	st, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.SetMode(mode); err != nil {
		t.Fatal(err)
	}
	skinStore, err := skins.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	info := runtimeinfo.Build("127.0.0.1", 19123, "test")
	return autoRankEnv{srv: server.New(st, skinStore, h, obsbridge.New(), info, webassets.FS, nil), st: st, h: h, dir: dir}
}

func rankRead(role, tier string, div int) owtracker.RankRead {
	return owtracker.RankRead{Role: role, Tier: tier, Division: div, Rank: owtracker.RankIndex(tier, div)}
}

// broadcasts drains the hub messages sent so far.
func broadcasts(t *testing.T, ch <-chan []byte) []hub.Message {
	t.Helper()
	var out []hub.Message
	for {
		select {
		case b := <-ch:
			var m hub.Message
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}
			out = append(out, m)
		case <-time.After(50 * time.Millisecond):
			return out
		}
	}
}

func TestApplyAutoRankClassic(t *testing.T) {
	env := newAutoRankEnv(t, store.ModeClassic)
	ch := env.h.Subscribe()
	env.srv.ApplyHotkeyFromAuto("win")
	env.srv.ApplyAutoRank(rankRead(store.RoleDamage, "Emerald", 2))

	st := env.st.State()
	if st.Wins != 1 || st.Rank != owtracker.RankIndex("Emerald", 2) {
		t.Fatalf("state: wins=%d rank=%d", st.Wins, st.Rank)
	}
	msgs := broadcasts(t, ch)
	if len(msgs) != 2 || msgs[1].Type != "rank" || msgs[1].View.Rank != st.Rank {
		t.Fatalf("broadcasts: %+v", msgs)
	}
}

func TestApplyAutoRankSplitMovesResultToRolePlayed(t *testing.T) {
	env := newAutoRankEnv(t, store.ModeRolesSplit)
	if _, err := env.st.SetRole(store.RoleTank); err != nil {
		t.Fatal(err)
	}
	ch := env.h.Subscribe()
	env.srv.ApplyHotkeyFromAuto("loss") // counted on the selected role (tank)
	env.srv.ApplyAutoRank(rankRead(store.RoleDamage, "Diamond", 4))

	st := env.st.State()
	tank, dmg := st.Roles[store.RoleTank], st.Roles[store.RoleDamage]
	if tank.Losses != 0 || dmg.Losses != 1 {
		t.Errorf("loss not moved to damage: tank=%+v damage=%+v", tank, dmg)
	}
	if dmg.Rank != owtracker.RankIndex("Diamond", 4) || tank.Rank != 0 {
		t.Errorf("ranks: tank=%d damage=%d", tank.Rank, dmg.Rank)
	}
	if st.Role != store.RoleDamage {
		t.Errorf("shown role %s, want damage", st.Role)
	}
	msgs := broadcasts(t, ch)
	last := msgs[len(msgs)-1]
	if last.Type != "rank" || last.View.Role != store.RoleDamage || last.View.Losses != 1 {
		t.Errorf("overlay got %+v", last.View)
	}

	// A second read for the same match moves nothing again.
	env.srv.ApplyAutoRank(rankRead(store.RoleSupport, "Gold", 1))
	st = env.st.State()
	if st.Roles[store.RoleDamage].Losses != 1 || st.Roles[store.RoleSupport].Losses != 0 {
		t.Errorf("result moved twice: %+v", st.Roles)
	}
}

func TestApplyAutoRankManualResultNotMoved(t *testing.T) {
	env := newAutoRankEnv(t, store.ModeRolesSplit)
	if _, err := env.st.SetRole(store.RoleTank); err != nil {
		t.Fatal(err)
	}
	env.srv.ApplyHotkey("win") // streamer's own hotkey: never moved
	env.srv.ApplyAutoRank(rankRead(store.RoleSupport, "Gold", 3))
	st := env.st.State()
	if st.Roles[store.RoleTank].Wins != 1 || st.Roles[store.RoleSupport].Wins != 0 {
		t.Errorf("manual win moved: %+v", st.Roles)
	}
	if st.Roles[store.RoleSupport].Rank != owtracker.RankIndex("Gold", 3) {
		t.Errorf("support rank %d", st.Roles[store.RoleSupport].Rank)
	}
}

func TestApplyAutoRankIgnored(t *testing.T) {
	// Open queue in a role mode: nowhere to put it.
	env := newAutoRankEnv(t, store.ModeRolesShared)
	ch := env.h.Subscribe()
	before := env.st.State()
	env.srv.ApplyAutoRank(rankRead("open", "Master", 5))
	if after := env.st.State(); after.Rank != before.Rank || after.Roles[store.RoleTank].Rank != 0 || after.Role != before.Role {
		t.Errorf("open queue changed role-mode state: %+v", after)
	}
	if msgs := broadcasts(t, ch); len(msgs) != 0 {
		t.Errorf("broadcast for an ignored read: %+v", msgs)
	}

	// Apex: Overwatch ranks never touch it.
	apex := newAutoRankEnv(t, store.ModeClassic)
	if _, err := apex.st.SetGame(store.GameApex); err != nil {
		t.Fatal(err)
	}
	apex.srv.ApplyAutoRank(rankRead(store.RoleDamage, "Champion", 1))
	if r := apex.st.State().Rank; r != 0 {
		t.Errorf("apex rank changed to %d", r)
	}
}

func TestAutomationRankEndpoint(t *testing.T) {
	env := newAutoRankEnv(t, store.ModeClassic)
	env.srv.Auto = owtracker.New(env.dir, nil)
	ts := httptest.NewServer(env.srv.Handler())
	defer ts.Close()

	post := func(q string) owtracker.Status {
		t.Helper()
		res, err := http.Post(ts.URL+"/api/automation/rank"+q, "", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("POST %s: %s", q, res.Status)
		}
		var st owtracker.Status
		if err := json.NewDecoder(res.Body).Decode(&st); err != nil {
			t.Fatal(err)
		}
		return st
	}
	if st := post("?enabled=1"); !st.RankAuto {
		t.Fatal("enabled=1 did not turn auto rank on")
	}
	if !owtracker.New(env.dir, nil).RankAuto() {
		t.Fatal("switch not persisted across restarts")
	}
	if st := post(""); st.RankAuto {
		t.Fatal("POST without enabled did not toggle off")
	}
	if st := post("?enabled=0"); st.RankAuto {
		t.Fatal("enabled=0 left auto rank on")
	}
	res, err := http.Get(ts.URL + "/api/automation/rank?enabled=1")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed || env.srv.Auto.RankAuto() {
		t.Fatalf("GET must not switch it: %s", res.Status)
	}
}
