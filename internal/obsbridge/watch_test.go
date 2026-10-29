package obsbridge

import (
	"math"
	"testing"

	"github.com/andreykaipov/goobs/api/typedefs"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestPlanFoldMovesStretchIntoWidget(t *testing.T) {
	tr := typedefs.SceneItemTransform{ScaleX: 1.64, ScaleY: 1.64, SourceWidth: 400, SourceHeight: 120}
	var asked float64
	plan, ok := planFold(tr, func(f float64) float64 { asked = f; return f })
	if !ok {
		t.Fatal("expected a fold")
	}
	if !near(asked, 1.64) {
		t.Fatalf("asked factor %v, want 1.64", asked)
	}
	// Source grows by the stretch, OBS scale goes back to 100%: same size on screen.
	if plan.width != 656 || plan.height != 197 || !near(plan.scaleX, 1) || !near(plan.scaleY, 1) {
		t.Fatalf("got %+v", plan)
	}
}

func TestPlanFoldKeepsWhatTheLimitCutOff(t *testing.T) {
	tr := typedefs.SceneItemTransform{ScaleX: 4, ScaleY: 4, SourceWidth: 300, SourceHeight: 100}
	// The widget could only grow 2x (scale limit); OBS keeps the other 2x.
	plan, ok := planFold(tr, func(float64) float64 { return 2 })
	if !ok || plan.width != 600 || plan.height != 200 || !near(plan.scaleX, 2) || !near(plan.scaleY, 2) {
		t.Fatalf("got %+v ok=%v", plan, ok)
	}
}

func TestPlanFoldIgnoresNoStretch(t *testing.T) {
	called := false
	onStretch := func(float64) float64 { called = true; return 1 }
	if _, ok := planFold(typedefs.SceneItemTransform{ScaleX: 1.01, ScaleY: 0.995, SourceWidth: 300, SourceHeight: 100}, onStretch); ok || called {
		t.Fatal("a near-100% scale must be left alone")
	}
	if _, ok := planFold(typedefs.SceneItemTransform{ScaleX: 2, ScaleY: 2}, onStretch); ok {
		t.Fatal("unknown source size must be left alone")
	}
	// At the limit already: nothing applied, nothing changed.
	if _, ok := planFold(typedefs.SceneItemTransform{ScaleX: 2, ScaleY: 2, SourceWidth: 300, SourceHeight: 100}, onStretch); ok {
		t.Fatal("a stretch the widget did not take must not touch OBS")
	}
}

func TestPlanFoldUnevenStretchUsesAverage(t *testing.T) {
	tr := typedefs.SceneItemTransform{ScaleX: 2, ScaleY: 0.5 * 2, SourceWidth: 200, SourceHeight: 100}
	plan, ok := planFold(tr, func(f float64) float64 { return f })
	want := math.Sqrt(2)
	if !ok || !near(plan.scaleX, 2/want) || !near(plan.scaleY, 1/want) {
		t.Fatalf("got %+v", plan)
	}
}

func TestGrowToNeverShrinksOutsideALookChange(t *testing.T) {
	cur := [2]int{768, 193}
	// start-up measure before the font loaded: narrower, must not cut the widget
	if got := growTo(cur, [2]int{648, 193}, false); got != cur {
		t.Fatalf("grow-only shrank the source: %v", got)
	}
	// longer rank name: the source grows
	if got := growTo(cur, [2]int{802, 190}, false); got != [2]int{802, 193} {
		t.Fatalf("got %v", got)
	}
	// new look: exactly the widget size, smaller or not
	if got := growTo(cur, [2]int{300, 120}, true); got != [2]int{300, 120} {
		t.Fatalf("got %v", got)
	}
	// unknown source size: take the widget size
	if got := growTo([2]int{}, [2]int{400, 100}, false); got != [2]int{400, 100} {
		t.Fatalf("got %v", got)
	}
}
