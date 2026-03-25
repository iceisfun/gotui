package layout

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
)

// Modal is a centered floating overlay that, when visible, renders its child
// above the normal renderable tree and consumes all input events. When hidden
// it renders nothing and passes events through.
//
// Modal implements render.Renderable, render.Overlayable, and render.Interactive.
type Modal struct {
	Child   render.Renderable
	visible bool
	width   int
	height  int

	// Screen dimensions, stored from the most recent Layout or set explicitly.
	screenW int
	screenH int
}

// NewModal creates a Modal with the given child and dimensions.
func NewModal(child render.Renderable, width, height int) *Modal {
	return &Modal{
		Child:  child,
		width:  width,
		height: height,
	}
}

// Show makes the modal visible.
func (m *Modal) Show() { m.visible = true }

// Hide makes the modal invisible.
func (m *Modal) Hide() { m.visible = false }

// IsVisible reports whether the modal is currently shown.
func (m *Modal) IsVisible() bool { return m.visible }

// SetScreenSize tells the modal the current terminal dimensions so it can
// center itself. This is also called automatically if the modal receives a
// Layout call.
func (m *Modal) SetScreenSize(w, h int) {
	m.screenW = w
	m.screenH = h
}

// Layout stores the available bounds so the modal can center itself.
func (m *Modal) Layout(bounds render.Rect) {
	m.screenW = bounds.Width
	m.screenH = bounds.Height
}

// Render draws nothing — the modal's child is rendered via the overlay system.
func (m *Modal) Render(v *render.View) {}

// Overlays returns an OverlayRequest that positions the child centered on
// screen when visible. Returns nil when hidden.
func (m *Modal) Overlays() []render.OverlayRequest {
	if !m.visible || m.Child == nil {
		return nil
	}

	w := m.width
	h := m.height

	// Center on screen.
	x := (m.screenW - w) / 2
	y := (m.screenH - h) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return []render.OverlayRequest{
		{
			Renderable: m.Child,
			Anchor:     render.Rect{X: x, Y: y, Width: w, Height: h},
			ZOrder:     100, // High z-order so it renders above most overlays.
		},
	}
}

// HandleEvent consumes ALL events when the modal is visible, forwarding
// keyboard/paste events to the child. When hidden it returns false so events
// pass through normally.
func (m *Modal) HandleEvent(ev input.Event) bool {
	if !m.visible {
		return false
	}

	// Forward keyboard and paste events to the child.
	if ev.Type == input.EventKey || ev.Type == input.EventPaste {
		if h, ok := m.Child.(render.Interactive); ok {
			h.HandleEvent(ev)
		}
	}

	// Consume all events when visible (including mouse) to prevent
	// interaction with content behind the modal.
	return true
}
