package layout

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
)

// AbsoluteChild describes a child positioned at an absolute location.
type AbsoluteChild struct {
	Renderable render.Renderable
	Rect       render.Rect // Position and size within the parent.
}

// Absolute positions children at explicit coordinates.
// Useful for floating panels, status bars at fixed positions, etc.
type Absolute struct {
	items  []AbsoluteChild
	bounds render.Rect
}

// NewAbsolute creates an absolute positioning container.
func NewAbsolute(children ...AbsoluteChild) *Absolute {
	return &Absolute{items: children}
}

// Layout stores the bounds and recursively layouts children.
func (a *Absolute) Layout(bounds render.Rect) {
	a.bounds = bounds
	for _, c := range a.items {
		if l, ok := c.Renderable.(interface{ Layout(render.Rect) }); ok {
			l.Layout(c.Rect)
		}
	}
}

// Render renders all children at their absolute positions, clipped to bounds.
func (a *Absolute) Render(v *render.View) {
	for _, c := range a.items {
		clipped := c.Rect.Intersect(render.Rect{
			X: 0, Y: 0,
			Width: a.bounds.Width, Height: a.bounds.Height,
		})
		if clipped.IsEmpty() {
			continue
		}
		sub := v.Sub(clipped)
		c.Renderable.Render(sub)
	}
}

// HandleEvent dispatches to children (last child = top, checked first).
func (a *Absolute) HandleEvent(ev input.Event) bool {
	for i := len(a.items) - 1; i >= 0; i-- {
		c := a.items[i]
		if h, ok := c.Renderable.(render.Interactive); ok {
			if h.HandleEvent(ev) {
				return true
			}
		}
	}
	return false
}

// Children returns all child renderables.
func (a *Absolute) Children() []render.Renderable {
	rs := make([]render.Renderable, len(a.items))
	for i, c := range a.items {
		rs[i] = c.Renderable
	}
	return rs
}

// ChildBounds returns the bounds of each child.
func (a *Absolute) ChildBounds() []render.Rect {
	rs := make([]render.Rect, len(a.items))
	for i, c := range a.items {
		rs[i] = c.Rect
	}
	return rs
}
