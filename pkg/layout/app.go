package layout

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/term"
)

// App is the top-level event loop owner. It manages the terminal, layout tree,
// event dispatch, focus, global key bindings, and compositing.
type App struct {
	term    *term.Terminal
	root    render.Renderable
	decoder *input.Decoder

	// Double-buffered composited output.
	front *render.Buffer
	back  *render.Buffer

	// Layer buffers for compositing.
	baseBuf    *render.Buffer
	overlayBuf *render.Buffer

	// Focus management.
	focused render.Focusable

	// Global key bindings — checked before focused renderable.
	bindings []render.KeyBinding

	// Set true to request a re-render.
	dirty bool

	// Quit flag — checked after dispatching events.
	quitting bool

	cols, rows int
}

// NewApp creates an App with the given terminal and root renderable.
func NewApp(t *term.Terminal, root render.Renderable) *App {
	cols, rows, _ := t.Size()
	return &App{
		term:       t,
		root:       root,
		decoder:    input.NewDecoder(),
		front:      render.NewBuffer(cols, rows),
		back:       render.NewBuffer(cols, rows),
		baseBuf:    render.NewBuffer(cols, rows),
		overlayBuf: render.NewBuffer(cols, rows),
		cols:       cols,
		rows:       rows,
	}
}

// SetFocus sets the currently focused renderable.
func (a *App) SetFocus(f render.Focusable) {
	if a.focused != nil {
		a.focused.Blur()
	}
	a.focused = f
	if f != nil {
		f.Focus()
	}
	a.dirty = true
}

// Bind adds a global key binding.
func (a *App) Bind(b render.KeyBinding) {
	a.bindings = append(a.bindings, b)
}

// BindKey is a convenience for binding a special key.
func (a *App) BindKey(key input.KeyCode, mod input.ModMask, action func()) {
	a.Bind(render.KeyBinding{Key: key, Mod: mod, Action: action})
}

// BindRune is a convenience for binding a rune with modifiers.
func (a *App) BindRune(r rune, mod input.ModMask, action func()) {
	a.Bind(render.KeyBinding{Key: input.KeyRune, Mod: mod, Rune: r, Action: action})
}

// RequestRender marks the screen as needing a redraw.
func (a *App) RequestRender() {
	a.dirty = true
}

// Quit signals the event loop to exit after the current dispatch cycle.
func (a *App) Quit() {
	a.quitting = true
}

// Root returns the root renderable.
func (a *App) Root() render.Renderable {
	return a.root
}

// Run starts the main event loop. It blocks until Quit is called or ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	// Handle SIGWINCH.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	defer signal.Stop(sigCh)

	a.term.StartReader(ctx)

	// Initial layout and render.
	a.resize()
	a.render()

	var batch []input.Event

	for {
		batch = batch[:0]

		// Block until at least one raw chunk arrives.
		select {
		case raw, ok := <-a.term.Reader().Events():
			if !ok {
				return nil
			}
			batch = append(batch, a.decoder.Feed(raw)...)

		case <-sigCh:
			a.resize()

		case <-ctx.Done():
			return ctx.Err()
		}

		// Drain all pending raw chunks before processing — prevents
		// input backlog when events arrive faster than we can render.
	drain:
		for {
			select {
			case raw, ok := <-a.term.Reader().Events():
				if !ok {
					return nil
				}
				batch = append(batch, a.decoder.Feed(raw)...)
			case <-sigCh:
				a.resize()
			default:
				break drain
			}
		}

		// Coalesce: collapse consecutive mouse-motion events into the last one.
		batch = coalesceEvents(batch)

		for _, ev := range batch {
			a.dispatch(ev)
			if a.quitting {
				return nil
			}
		}

		if a.dirty {
			a.render()
		}
	}
}

func (a *App) resize() {
	cols, rows, err := a.term.Size()
	if err != nil {
		return
	}
	a.cols = cols
	a.rows = rows
	a.front.Resize(cols, rows)
	a.front.Clear()
	a.back.Resize(cols, rows)
	a.baseBuf.Resize(cols, rows)
	a.overlayBuf.Resize(cols, rows)

	// Re-layout the root.
	if c, ok := a.root.(interface{ Layout(render.Rect) }); ok {
		c.Layout(render.Rect{X: 0, Y: 0, Width: cols, Height: rows})
	}

	a.dirty = true
}

func (a *App) render() {
	a.dirty = false

	// Render root into base buffer.
	a.baseBuf.Clear()
	baseView := render.NewView(a.baseBuf, render.Rect{X: 0, Y: 0, Width: a.cols, Height: a.rows})
	a.root.Render(baseView)

	// Collect and render overlays.
	overlays := CollectOverlays(a.root, render.Rect{X: 0, Y: 0, Width: a.cols, Height: a.rows})
	if len(overlays) > 0 {
		SortOverlays(overlays)
		a.overlayBuf.ClearTransparent()
		for _, o := range overlays {
			if o.bounds.IsEmpty() {
				continue
			}
			sub := render.NewView(a.overlayBuf, o.bounds)
			o.renderable.Render(sub)
		}
		// Composite base + overlay into back buffer.
		render.CompositeTwo(a.back, a.baseBuf, a.overlayBuf)
	} else {
		// No overlays — copy base directly.
		render.CompositeTwo(a.back, a.baseBuf, a.baseBuf)
	}

	// Flush diff to terminal.
	render.Flush(a.term.Writer(), a.front, a.back)

	// Position hardware cursor if focused renderable has one.
	if cur, ok := a.focused.(render.Cursorable); ok {
		cx, cy, visible := cur.CursorPosition()
		if visible {
			// Translate to screen coordinates.
			// Find the focused renderable's screen position.
			ox, oy := a.findScreenOffset(a.focused)
			a.term.Writer().MoveTo(oy+cy+1, ox+cx+1) // 1-based
			a.term.Writer().ShowCursor()
		} else {
			a.term.Writer().HideCursor()
		}
	} else {
		a.term.Writer().HideCursor()
	}

	a.term.Writer().Flush()
}

func (a *App) dispatch(ev input.Event) {
	a.dirty = true

	// Resize events.
	if ev.Type == input.EventResize {
		a.resize()
		return
	}

	// Global key bindings (checked first).
	if ev.Type == input.EventKey {
		for _, b := range a.bindings {
			if b.MatchEvent(ev.Key) {
				b.Action()
				return
			}
		}
	}

	// Mouse events: hit-test and dispatch to target.
	if ev.Type == input.EventMouse {
		a.dispatchMouse(ev)
		return
	}

	// Keyboard/paste: dispatch to focused renderable.
	if a.focused != nil {
		if a.focused.HandleEvent(ev) {
			return
		}
	}

	// Bubble up through root if it's Interactive.
	if h, ok := a.root.(render.Interactive); ok {
		h.HandleEvent(ev)
	}
}

func (a *App) dispatchMouse(ev input.Event) {
	sx, sy := ev.Mouse.ScreenX, ev.Mouse.ScreenY

	// 1. Check overlays first (topmost = highest z-order, checked first).
	rootBounds := render.Rect{X: 0, Y: 0, Width: a.cols, Height: a.rows}
	overlays := CollectOverlays(a.root, rootBounds)
	if len(overlays) > 0 {
		SortOverlays(overlays)
		// Walk from highest z-order to lowest.
		for i := len(overlays) - 1; i >= 0; i-- {
			o := overlays[i]
			if o.bounds.Contains(sx, sy) {
				localEv := ev
				localEv.Mouse.X = sx - o.bounds.X
				localEv.Mouse.Y = sy - o.bounds.Y
				if h, ok := o.renderable.(render.Interactive); ok {
					h.HandleEvent(localEv)
				}
				return
			}
		}
	}

	// 2. Hit-test the container tree.
	target, bounds := a.hitTest(a.root, rootBounds, sx, sy)
	if target == nil {
		return
	}

	// Set local coordinates relative to the target's bounds.
	localEv := ev
	localEv.Mouse.X = sx - bounds.X
	localEv.Mouse.Y = sy - bounds.Y

	// Click on a Focusable shifts focus.
	if ev.Mouse.Button == input.MouseLeft {
		if f, ok := target.(render.Focusable); ok {
			if a.focused != f {
				a.SetFocus(f)
			}
		}
	}

	// Dispatch to target.
	if h, ok := target.(render.Interactive); ok {
		h.HandleEvent(localEv)
	}
}

// hitTest walks the tree to find the deepest renderable containing (x, y).
func (a *App) hitTest(r render.Renderable, bounds render.Rect, x, y int) (render.Renderable, render.Rect) {
	if !bounds.Contains(x, y) {
		return nil, render.Rect{}
	}

	// Check container children (deepest match wins).
	if c, ok := r.(render.Container); ok {
		children := c.Children()
		childBounds := c.ChildBounds()
		// Check in reverse order (last = topmost).
		for i := len(children) - 1; i >= 0; i-- {
			if i < len(childBounds) {
				if hit, hb := a.hitTest(children[i], childBounds[i], x, y); hit != nil {
					return hit, hb
				}
			}
		}
	}

	// No child matched — return self if it handles interaction.
	if _, ok := r.(render.Interactive); ok {
		return r, bounds
	}

	return nil, render.Rect{}
}

// findScreenOffset locates a renderable's absolute screen position by walking
// the tree from root.
func (a *App) findScreenOffset(target render.Renderable) (int, int) {
	ox, oy, _ := a.findOffset(a.root, render.Rect{X: 0, Y: 0, Width: a.cols, Height: a.rows}, target)
	return ox, oy
}

func (a *App) findOffset(r render.Renderable, bounds render.Rect, target render.Renderable) (int, int, bool) {
	if r == target {
		return bounds.X, bounds.Y, true
	}
	if c, ok := r.(render.Container); ok {
		children := c.Children()
		childBounds := c.ChildBounds()
		for i, child := range children {
			if i < len(childBounds) {
				if ox, oy, found := a.findOffset(child, childBounds[i], target); found {
					return ox, oy, true
				}
			}
		}
	}
	return 0, 0, false
}

// coalesceEvents collapses consecutive mouse-motion events into the last one.
// This prevents a flood of motion events from building up render backlog.
// Non-motion events (clicks, keys, paste) are always preserved.
func coalesceEvents(events []input.Event) []input.Event {
	if len(events) <= 1 {
		return events
	}

	result := events[:0]
	for i, ev := range events {
		if ev.Type == input.EventMouse && ev.Mouse.Button == input.MouseMotion {
			// Skip this motion event if the next event is also a motion event.
			if i+1 < len(events) {
				next := events[i+1]
				if next.Type == input.EventMouse && next.Mouse.Button == input.MouseMotion {
					continue
				}
			}
		}
		result = append(result, ev)
	}
	return result
}
