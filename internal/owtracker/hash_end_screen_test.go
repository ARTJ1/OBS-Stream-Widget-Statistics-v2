package owtracker

import "testing"

func TestIsHashEndScreen(t *testing.T) {
	if !isHashEndScreen(20, 36, 5, 8, 8) {
		t.Fatal("expected hash end screen when lossDist within cap")
	}
	if isHashEndScreen(45, 50, 2, 3, 8) {
		t.Fatal("expected weak gameplay hash to be rejected")
	}
	if !isHashEndScreen(50, 50, 58, 10, 8) {
		t.Fatal("expected pixel match to open hash gate")
	}
}
