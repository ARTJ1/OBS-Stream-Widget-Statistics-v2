package owtracker

import (
	"image/color"
	"testing"
)

func TestBuiltinTemplates(t *testing.T) {
	if err := verifyDefaultTemplates(); err != nil {
		t.Fatal(err)
	}
}

func TestNewUsesBuiltinDefaults(t *testing.T) {
	tr := New(t.TempDir(), nil)
	st := tr.Status()
	if !st.WinReady || !st.LossReady || !st.ZoneReady {
		t.Fatalf("expected builtin ready, got %+v", st)
	}
	if st.WinCustom || st.LossCustom {
		t.Fatalf("expected builtin templates, got custom flags %+v", st)
	}
	win, loss := tr.hashes.get()
	if win != defaultWinHash || loss != defaultLossHash {
		t.Fatalf("hashes win=%s loss=%s", win, loss)
	}
}

func TestResetToDefaults(t *testing.T) {
	dir := t.TempDir()
	tr := New(dir, nil)
	img := solidRGBA(480, 180, color.RGBA{R: 200, G: 160, B: 30, A: 255})
	h, err := computeTemplateHash(img)
	if err != nil {
		t.Fatal(err)
	}
	tr.hashes.setWin(h)
	saved := savedHashes{Win: h, CustomWin: true, MatchThreshold: 10}
	if err := saveHashes(dir, saved); err != nil {
		t.Fatal(err)
	}
	if err := tr.ResetToDefaults(); err != nil {
		t.Fatal(err)
	}
	st := tr.Status()
	win, _ := tr.hashes.get()
	if st.WinCustom || win == h {
		t.Fatalf("expected revert to builtin, st=%+v win=%s", st, win)
	}
}
