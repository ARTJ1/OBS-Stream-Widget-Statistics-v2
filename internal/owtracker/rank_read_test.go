package owtracker

import (
	"testing"
	"time"
)

func TestMatchRankRole(t *testing.T) {
	cases := map[string]string{
		"lll УРОН":         "damage",
		"Н! УРОН":          "damage",
		"урон":             "damage",
		"111 DAMAGE":       "damage",
		"+ SUPPORT":        "support",
		"ПОДДЕРЖКА":        "support",
		"ТАНК":             "tank",
		"TANK":             "tank",
		"TAMK":             "tank", // one misread letter in a 4-letter word
		"TAHK":             "tank", // ТАНК read in Latin look-alikes
		"YPOH":             "damage",
		"n ano nonuuo":     "", // gameplay text: "nano" is not DANO
		"* TANK":           "tank",
		"OPEN QUEUE":       RankRoleOpen,
		"DAMAGE MITIGATED": "damage", // matched word; the label check rejects the frame
		"COMPETITIVE":      "",
		"ПОХВАЛИТЬ":        "",
		"":                 "",
	}
	for in, want := range cases {
		if got := matchRankRole(in); got != want {
			t.Errorf("matchRankRole(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatchRankLabel(t *testing.T) {
	cases := []struct {
		in   string
		tier string
		div  int
	}{
		{"ИЗУМРУД 2", "Emerald", 2},
		{"EMERALD 3", "Emerald", 3},
		{"DIAMOND 5", "Diamond", 5},
		{"DIRMOND5", "Diamond", 5},
		{"ППЯТИНЯ 2", "Platinum", 2},
		{"ИЗУМРУД З", "Emerald", 3},
		{"DIAMOND I", "Diamond", 1},
		{"ЗОЛОТО", "", 0}, // no division, the З belongs to the word
		{"ПЛАТИНА 2", "Platinum", 2},
		{"fMfRßlD3", "Emerald", 3},
		{"АЛМАЗ 4", "Diamond", 4},
		{"GRANDMASTER 1", "Grandmaster", 1},
		{"MASTER 2", "Master", 2},
		{"ГРАНДМАСТЕР 3", "Grandmaster", 3},
		{"CHAMPION 5", "Champion", 5},
		{"НЗУМРУД?", "", 0},  // division unreadable
		{"muon 5", "", 0},    // too far from any tier
		{"DIAMOND 7", "", 0}, // no such division
		{"RANK PROGRESS: 7%", "", 0},
		// EN client: 4-letter GOLD and the division 1 read as a letter.
		{"GOLD I", "Gold", 1},
		{"GOLDI", "Gold", 1},
		{"GOLD 1", "Gold", 1},
		{"G0LD 2", "Gold", 2},
		{"COLD 4", "", 0}, // a wrong first letter is too loose for a 4-letter word
		{"SILVER l", "Silver", 1},
		{"PLATINUM S", "Platinum", 5},
		{"BRONZE 5", "Bronze", 5},
		{"ИЗУМРУДЗ", "Emerald", 3},
		{"АЛМАЗ", "", 0}, // ends in a З-like letter but has no division
		{"GOLD", "", 0},
		{"TOP 500", "", 0},
		{"EMERALD 3%", "Emerald", 3},
		// The italic 4 read as a letter, Latin look-alikes, Л as two strokes.
		{"ГРАНДМАСТЕР а", "Grandmaster", 4},
		{"EMERALD a", "Emerald", 4},
		{"ПЛАТИНА Ч", "Platinum", 4},
		{"AJlMA32", "Diamond", 2},
		{"3OЛOTO 5", "Gold", 5}, // Latin O and a digit 3 in a Cyrillic word
		{"ПЛАТИНА", "", 0},      // the division was not read: no guess
		{"SILVER L", "", 0},     // an L is not a safe 1 (this was a dark "SILVER 4")
	}
	for _, c := range cases {
		tier, div := matchRankLabel(c.in)
		if tier != c.tier || div != c.div {
			t.Errorf("matchRankLabel(%q) = %q %d, want %q %d", c.in, tier, div, c.tier, c.div)
		}
	}
}

func TestRankIndex(t *testing.T) {
	if got := RankIndex("Bronze", 5); got != 0 {
		t.Errorf("Bronze 5 = %d", got)
	}
	if got := RankIndex("Emerald", 2); got != 4*5+3 {
		t.Errorf("Emerald 2 = %d", got)
	}
	if got := RankIndex("Champion", 1); got != 44 {
		t.Errorf("Champion 1 = %d", got)
	}
	if got := RankIndex("Diamond", 0); got != -1 {
		t.Errorf("Diamond 0 = %d", got)
	}
}

func TestRankWatch(t *testing.T) {
	now := time.Now()
	at := func(d time.Duration) time.Time { return now.Add(d) }
	em3 := RankRead{Role: "damage", Tier: "Emerald", Division: 3, Rank: RankIndex("Emerald", 3)}
	animating := RankRead{Role: "damage"}
	var w rankWatch

	// Two equal reads do not commit while the screen is up; the screen closing does.
	w.start(now)
	w.observe(at(0), animating)
	w.observe(at(250*time.Millisecond), em3)
	if _, ok := w.observe(at(500*time.Millisecond), em3); ok {
		t.Fatal("committed with the screen still up")
	}
	if _, ok := w.observe(at(time.Second), RankRead{}); ok {
		t.Fatal("committed before the screen was gone for rankGoneAfter")
	}
	if got, ok := w.observe(at(500*time.Millisecond+rankGoneAfter), RankRead{}); !ok || got != em3 {
		t.Fatalf("screen closed: %+v %v", got, ok)
	}
	if w.active() {
		t.Fatal("still active after commit")
	}

	// A single read commits once the screen is gone.
	w.start(now)
	w.observe(at(0), em3)
	if got, ok := w.observe(at(rankGoneAfter), RankRead{}); !ok || got != em3 {
		t.Fatalf("lone read: %+v %v", got, ok)
	}

	// Nothing read: the watch ends after the search window with no rank.
	w.start(now)
	if _, ok := w.observe(at(rankSearchFor-time.Second), RankRead{}); ok || !w.active() {
		t.Fatal("ended before the search window")
	}
	if _, ok := w.observe(at(rankSearchFor), RankRead{}); ok {
		t.Fatal("rank from nothing")
	}
	if w.active() {
		t.Fatal("still active after the search window")
	}
}

// A rank-up left on screen: the old rank shows first, then the new one (seen
// live: Emerald 2 -> Emerald 1 read as Emerald 2).
func TestRankWatchRankUp(t *testing.T) {
	now := time.Now()
	em2 := RankRead{Role: "damage", Tier: "Emerald", Division: 2, Rank: RankIndex("Emerald", 2)}
	em1 := RankRead{Role: "damage", Tier: "Emerald", Division: 1, Rank: RankIndex("Emerald", 1)}
	var w rankWatch
	w.start(now)
	tick := time.Duration(0)
	feed := func(r RankRead, n int) (RankRead, bool) {
		for i := 0; i < n; i++ {
			if got, ok := w.observe(now.Add(40*time.Second+tick), r); ok {
				return got, ok
			}
			tick += rankPoll
		}
		return RankRead{}, false
	}
	if got, ok := feed(em2, 16); ok { // bar filling, 4 s
		t.Fatalf("committed the old rank %+v", got)
	}
	if got, ok := feed(RankRead{Role: "damage"}, 6); ok { // rank-up flash
		t.Fatalf("committed during the animation: %+v", got)
	}
	// The new rank stays on screen: committed after rankSettle.
	got, ok := feed(em1, int(rankSettle/rankPoll)+2)
	if !ok || got != em1 {
		t.Fatalf("got %+v %v, want %+v", got, ok, em1)
	}
	if tick > 4*time.Second+6*rankPoll+rankSettle+rankPoll {
		t.Errorf("committed late: %s after the screen showed", tick)
	}
}

// The screen shows late in the search window and stays: followed past it.
func TestRankWatchPastSearchWindow(t *testing.T) {
	now := time.Now()
	em1 := RankRead{Role: "tank", Tier: "Gold", Division: 1, Rank: RankIndex("Gold", 1)}
	var w rankWatch
	w.start(now)
	d := rankSearchFor - time.Second
	for ; d < rankSearchFor+5*time.Second; d += rankPoll {
		if got, ok := w.observe(now.Add(d), em1); ok {
			t.Fatalf("committed at %s: %+v", d, got)
		}
	}
	if got, ok := w.observe(now.Add(d+rankGoneAfter), RankRead{}); !ok || got != em1 {
		t.Fatalf("got %+v %v", got, ok)
	}

	// A screen that never closes and never settles (label always animating)
	// still ends at rankWatchMax.
	w.start(now)
	w.observe(now.Add(time.Minute), em1)
	for d := time.Minute; d < rankWatchMax; d += time.Second {
		w.observe(now.Add(d), RankRead{Role: "tank"})
	}
	if got, ok := w.observe(now.Add(rankWatchMax), RankRead{Role: "tank"}); !ok || got != em1 {
		t.Fatalf("at the cap: %+v %v", got, ok)
	}
}
