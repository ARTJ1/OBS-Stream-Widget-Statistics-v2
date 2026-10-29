package obsbridge

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/events/subscriptions"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/typedefs"
)

// Watcher keeps one quiet connection to OBS for two jobs:
//   - the Browser Source always matches the widget's size (SetSize);
//   - when the streamer stretches the source in OBS, the stretch moves into the
//     widget's own scale and the source goes back to 100%, so OBS never scales a
//     rendered picture (which is what makes text soapy).
//
// It only reacts to OBS events; nothing is polled while connected.
type Watcher struct {
	bridge *Bridge
	// OnStretch gets the stretch factor and returns the factor it actually
	// applied to the widget scale (after limits); 1 means nothing changed.
	OnStretch func(factor float64) float64

	mu      sync.Mutex
	client  *goobs.Client
	timer   *time.Timer
	pending [2]int // last widget size reported by the overlay
	exact   bool   // pending is for a new look: fit it exactly (else only grow)
	obsSize [2]int // the Browser Source size in OBS right now
	quiet   time.Time
	fold    *pendingFold // a stretch waiting for the widget to hide and report its new size
}

// pendingFold is the OBS side of a stretch, applied together with the new source size
// once the widget has hidden itself, so the swap is never seen on stream.
type pendingFold struct {
	scene string
	item  int
	t     typedefs.SceneItemTransform
	size  [2]int
	timer *time.Timer
}

const (
	foldWait       = 3 * time.Second        // the widget usually reports within ~1 s; then we go ahead anyway
	resizeRedraw   = 160 * time.Millisecond // a resized Browser Source needs a few frames to repaint
	watchRetry     = 5 * time.Second
	exitQuiet      = 30 * time.Second       // OBS still accepts connections while it closes
	stretchSettle  = 700 * time.Millisecond // wait until the streamer lets go of the mouse
	stretchEpsilon = 0.02
)

func NewWatcher(b *Bridge, onStretch func(float64) float64) *Watcher {
	return &Watcher{bridge: b, OnStretch: onStretch}
}

// Run connects whenever OBS is up and a scene is configured; returns when ctx ends.
func (w *Watcher) Run(ctx context.Context) {
	for {
		if cfg := w.bridge.Config(); cfg.Scene != "" && time.Now().After(w.quietUntil()) {
			if client, err := w.connect(cfg); err == nil {
				w.listen(ctx, client)
			}
		}
		select {
		case <-ctx.Done():
			w.drop()
			return
		case <-time.After(watchRetry):
		}
	}
}

func (w *Watcher) connect(cfg Config) (*goobs.Client, error) {
	opts := []goobs.Option{goobs.WithEventSubscriptions(
		subscriptions.General | subscriptions.SceneItems | subscriptions.SceneItemTransformChanged)}
	if cfg.Password != "" {
		opts = append(opts, goobs.WithPassword(cfg.Password))
	}
	client, err := goobs.New(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), opts...)
	if err != nil {
		return nil, err
	}
	// Start from the size OBS kept from last time; never shrink it on start-up.
	cur := [2]int{}
	if in, err := client.Inputs.GetInputSettings(inputs.NewGetInputSettingsParams().WithInputName(cfg.SourceName)); err == nil {
		if v, ok := in.InputSettings["width"].(float64); ok {
			cur[0] = int(v)
		}
		if v, ok := in.InputSettings["height"].(float64); ok {
			cur[1] = int(v)
		}
	}
	w.mu.Lock()
	w.client = client
	w.obsSize = cur
	pending, exact := w.pending, w.exact
	w.mu.Unlock()
	if pending[0] > 0 {
		w.SetSize(pending[0], pending[1], exact)
	}
	// A source stretched before we were running (or before this version) turns sharp too.
	if item, err := client.SceneItems.GetSceneItemId(sceneitems.NewGetSceneItemIdParams().
		WithSceneName(cfg.Scene).WithSourceName(cfg.SourceName)); err == nil {
		go w.foldStretch(cfg.Scene, item.SceneItemId)
	}
	return client, nil
}

func (w *Watcher) listen(ctx context.Context, client *goobs.Client) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		client.Listen(func(ev any) {
			switch e := ev.(type) {
			case *events.ExitStarted:
				// Let go at once so OBS never waits on us while closing.
				w.mu.Lock()
				w.quiet = time.Now().Add(exitQuiet)
				w.mu.Unlock()
				go w.drop()
			case *events.SceneItemTransformChanged:
				w.onTransform(e)
			}
		})
	}()
	select {
	case <-ctx.Done():
		w.drop()
		<-done
	case <-done:
		w.drop()
	}
}

func (w *Watcher) quietUntil() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.quiet
}

func (w *Watcher) drop() {
	w.mu.Lock()
	client := w.client
	w.client = nil
	if w.timer != nil {
		w.timer.Stop()
	}
	w.mu.Unlock()
	if client != nil {
		_ = client.Disconnect()
	}
}

func (w *Watcher) onTransform(e *events.SceneItemTransformChanged) {
	t := e.SceneItemTransform
	if t == nil || (math.Abs(t.ScaleX-1) < stretchEpsilon && math.Abs(t.ScaleY-1) < stretchEpsilon) {
		return
	}
	scene, id := e.SceneName, e.SceneItemId
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(stretchSettle, func() { w.foldStretch(scene, id) })
}

// foldStretch moves an OBS stretch of our source into the widget scale.
func (w *Watcher) foldStretch(scene string, itemID int) {
	w.mu.Lock()
	client := w.client
	w.mu.Unlock()
	if client == nil || w.OnStretch == nil {
		return
	}
	src, err := client.SceneItems.GetSceneItemSource(sceneitems.NewGetSceneItemSourceParams().
		WithSceneName(scene).WithSceneItemId(itemID))
	if err != nil || src.SourceName != w.bridge.Config().SourceName {
		return
	}
	cur, err := client.SceneItems.GetSceneItemTransform(sceneitems.NewGetSceneItemTransformParams().
		WithSceneName(scene).WithSceneItemId(itemID))
	if err != nil || cur.SceneItemTransform == nil {
		return
	}
	t := *cur.SceneItemTransform
	// A bounding box sizes the source on its own; leave that setup alone.
	if t.BoundsType != "" && t.BoundsType != "OBS_BOUNDS_NONE" {
		return
	}
	plan, ok := planFold(t, w.OnStretch)
	if !ok {
		return
	}
	t.ScaleX, t.ScaleY = plan.scaleX, plan.scaleY
	// OBS rejects bounds below 1 even when bounds are off.
	t.BoundsWidth = math.Max(t.BoundsWidth, 1)
	t.BoundsHeight = math.Max(t.BoundsHeight, 1)
	// The new scale went out to the widget (OnStretch); it hides, measures itself and
	// reports the size (SetSize), and that is when OBS changes. If it never reports
	// (source hidden, page not loaded), fall back to the planned size.
	f := &pendingFold{scene: scene, item: itemID, t: t, size: [2]int{plan.width, plan.height}}
	w.mu.Lock()
	if w.fold != nil && w.fold.timer != nil {
		w.fold.timer.Stop()
	}
	w.fold = f
	f.timer = time.AfterFunc(foldWait, func() { w.applyFold(f, f.size) })
	w.mu.Unlock()
}

// applyFold resizes the source and resets its OBS scale in one go.
func (w *Watcher) applyFold(f *pendingFold, size [2]int) {
	w.mu.Lock()
	if w.fold != f {
		w.mu.Unlock()
		return
	}
	w.fold = nil
	f.timer.Stop()
	client := w.client
	w.mu.Unlock()
	if client == nil {
		return
	}
	_ = w.resizeOffAir(client, f.scene, f.item, size, &f.t)
}

type foldPlan struct {
	width, height  int
	scaleX, scaleY float64
}

// planFold works out the new source size and the OBS scale that is left after
// the widget took over the stretch. On-screen size stays the same.
func planFold(t typedefs.SceneItemTransform, onStretch func(float64) float64) (foldPlan, bool) {
	stretch := math.Sqrt(math.Abs(t.ScaleX * t.ScaleY))
	if stretch == 0 || math.Abs(stretch-1) < stretchEpsilon || t.SourceWidth <= 0 || t.SourceHeight <= 0 {
		return foldPlan{}, false
	}
	applied := onStretch(stretch)
	if applied <= 0 || math.Abs(applied-1) < 0.001 {
		return foldPlan{}, false
	}
	return foldPlan{
		width:  int(math.Round(t.SourceWidth * applied)),
		height: int(math.Round(t.SourceHeight * applied)),
		scaleX: t.ScaleX / applied,
		scaleY: t.ScaleY / applied,
	}, true
}

// SetSize fits the Browser Source to the widget (device pixels). exact is for a new
// look (layout, scale, font, icon style): the source takes exactly that size. Any
// other report (start-up, counters, rank names) can only grow the source, so it never
// jumps or gets cut while the look stays the same.
func (w *Watcher) SetSize(width, height int, exact bool) {
	if width <= 0 || height <= 0 {
		return
	}
	w.mu.Lock()
	w.pending = [2]int{width, height}
	w.exact = w.exact || exact
	client, cur, fit, fold := w.client, w.obsSize, w.exact, w.fold
	w.mu.Unlock()
	if client == nil {
		return
	}
	if fold != nil && exact {
		// The widget is hidden now: new size and OBS scale change together.
		w.applyFold(fold, [2]int{width, height})
		return
	}
	target := growTo(cur, [2]int{width, height}, fit)
	if target == cur {
		return
	}
	if fit {
		// A new look: the widget has hidden itself, so the OBS redraw can be hidden too.
		cfg := w.bridge.Config()
		if item, err := client.SceneItems.GetSceneItemId(sceneitems.NewGetSceneItemIdParams().
			WithSceneName(cfg.Scene).WithSourceName(cfg.SourceName)); err == nil {
			_ = w.resizeOffAir(client, cfg.Scene, item.SceneItemId, target, nil)
			return
		}
	}
	_ = w.setInputSize(client, target[0], target[1])
}

// resizeOffAir changes the source size (and, for a stretch, its transform) with the
// scene item switched off: OBS paints a resized Browser Source black for a frame or two
// until the page redraws, and an item that is off is not drawn at all. The widget is
// hidden at this point, so switching the item off is not visible. An item the streamer
// switched off stays off.
func (w *Watcher) resizeOffAir(client *goobs.Client, scene string, item int, size [2]int, t *typedefs.SceneItemTransform) error {
	on := false
	if st, err := client.SceneItems.GetSceneItemEnabled(sceneitems.NewGetSceneItemEnabledParams().
		WithSceneName(scene).WithSceneItemId(item)); err == nil {
		on = st.SceneItemEnabled
	}
	setOn := func(v bool) {
		_, _ = client.SceneItems.SetSceneItemEnabled(sceneitems.NewSetSceneItemEnabledParams().
			WithSceneName(scene).WithSceneItemId(item).WithSceneItemEnabled(v))
	}
	if on {
		setOn(false)
		defer func() {
			time.Sleep(resizeRedraw)
			setOn(true)
		}()
	}
	if err := w.setInputSize(client, size[0], size[1]); err != nil {
		return err
	}
	if t != nil {
		_, _ = client.SceneItems.SetSceneItemTransform(sceneitems.NewSetSceneItemTransformParams().
			WithSceneName(scene).WithSceneItemId(item).WithSceneItemTransform(t))
	}
	return nil
}

// growTo is the source size to use: the widget size for a new look, otherwise the
// larger of the current source and the widget on each side.
func growTo(cur, widget [2]int, exact bool) [2]int {
	if exact || cur[0] <= 0 || cur[1] <= 0 {
		return widget
	}
	return [2]int{max(cur[0], widget[0]), max(cur[1], widget[1])}
}

func (w *Watcher) setInputSize(client *goobs.Client, width, height int) error {
	_, err := client.Inputs.SetInputSettings(inputs.NewSetInputSettingsParams().
		WithInputName(w.bridge.Config().SourceName).
		WithOverlay(true).
		WithInputSettings(map[string]any{"width": width, "height": height}))
	if err == nil {
		w.mu.Lock()
		w.obsSize = [2]int{width, height}
		w.exact = false
		w.mu.Unlock()
	}
	return err
}

// Size is the last widget size known to fit the Browser Source, if any.
func (w *Watcher) Size() (int, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.pending[0], w.pending[1]
}
