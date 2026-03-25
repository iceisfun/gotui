package layout

import (
	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
)

// Viewport wraps a Renderable whose content may be larger than its visible area.
// It manages scroll offset and renders only the visible portion.
type Viewport struct {
	Child   render.Renderable
	bounds  render.Rect
	offsetX int
	offsetY int
}

// NewViewport creates a scrolling viewport around a child renderable.
func NewViewport(child render.Renderable) *Viewport {
	return &Viewport{Child: child}
}

// Layout sets the visible area.
func (vp *Viewport) Layout(bounds render.Rect) {
	vp.bounds = bounds
}

// Render renders the child through the viewport's offset.
func (vp *Viewport) Render(v *render.View) {
	// The child renders into the full view. If it's Scrollable,
	// it should use ScrollOffset to shift its own content.
	if s, ok := vp.Child.(render.Scrollable); ok {
		s.SetScrollOffset(vp.offsetX, vp.offsetY)
	}
	vp.Child.Render(v)
}

// HandleEvent passes events to the child, handling scroll if not consumed.
func (vp *Viewport) HandleEvent(ev input.Event) bool {
	// Let the child handle first.
	if h, ok := vp.Child.(render.Interactive); ok {
		if h.HandleEvent(ev) {
			return true
		}
	}

	// Handle mouse wheel for scrolling.
	if ev.Type == input.EventMouse {
		switch ev.Mouse.Button {
		case input.MouseWheelUp:
			vp.ScrollBy(0, -3)
			return true
		case input.MouseWheelDown:
			vp.ScrollBy(0, 3)
			return true
		}
	}

	return false
}

// ScrollOffset returns the current scroll position.
func (vp *Viewport) ScrollOffset() (x, y int) {
	return vp.offsetX, vp.offsetY
}

// SetScrollOffset sets the scroll position.
func (vp *Viewport) SetScrollOffset(x, y int) {
	vp.offsetX = max(x, 0)
	vp.offsetY = max(y, 0)
}

// ScrollBy adjusts the scroll offset by a delta.
func (vp *Viewport) ScrollBy(dx, dy int) {
	vp.SetScrollOffset(vp.offsetX+dx, vp.offsetY+dy)
}

// ScrollTo ensures a given row is visible.
func (vp *Viewport) ScrollTo(row int) {
	if row < vp.offsetY {
		vp.offsetY = row
	} else if row >= vp.offsetY+vp.bounds.Height {
		vp.offsetY = row - vp.bounds.Height + 1
	}
}
