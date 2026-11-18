package store

import "testing"

func TestNeonLayoutSettings(t *testing.T) {
	s := newTestStore(t, ModeClassic)
	next := s.Settings()
	next.Layout = "neon"
	next.NeonColor = "bad"
	next.LayoutFx = "detailed-win,neon-breathe,neon-set"
	snap, err := s.UpdateSettings(next)
	if err != nil {
		t.Fatal(err)
	}
	got := snap.Settings
	if got.Layout != "neon" {
		t.Fatalf("layout = %q, want neon", got.Layout)
	}
	if got.NeonColor != DefaultNeonColor || got.NeonTextColor == "" || got.NeonWaveColor == "" {
		t.Fatalf("neon colours = %q %q %q", got.NeonColor, got.NeonTextColor, got.NeonWaveColor)
	}
	// A list that already knows the neon layout is kept as the user left it.
	if got.LayoutFx != "detailed-win,neon-breathe,neon-set" {
		t.Fatalf("layoutFx = %q", got.LayoutFx)
	}
}

func TestWithNeonFx(t *testing.T) {
	cases := map[string]string{
		"":                  neonLayoutFx,
		"none":              neonLayoutFx,
		"detailed-win":      "detailed-win," + neonLayoutFx,
		"mini-win,neon-set": "mini-win,neon-set",
	}
	for in, want := range cases {
		if got := withNeonFx(in); got != want {
			t.Errorf("withNeonFx(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNeonGlowClamp(t *testing.T) {
	for in, want := range map[int]int{-5: 0, 0: 0, 60: 60, 200: 200, 900: 200} {
		if got := clampNeonGlow(in); got != want {
			t.Errorf("clampNeonGlow(%d) = %d, want %d", in, got, want)
		}
	}
	if d := DefaultSettings(); d.NeonGlow != DefaultNeonGlow || d.NeonTextGlow != DefaultNeonGlow {
		t.Errorf("default glow = %d/%d, want %d", d.NeonGlow, d.NeonTextGlow, DefaultNeonGlow)
	}
}
