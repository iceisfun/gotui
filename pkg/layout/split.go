package layout

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
)

// Direction specifies the split axis.
type Direction uint8

const (
	Horizontal Direction = iota // Children laid out left to right.
	Vertical                    // Children laid out top to bottom.
)

// SplitChild describes one child in a split layout.
type SplitChild struct {
	Renderable render.Renderable
	// Size is the proportion of space this child occupies.
	// Values <= 1.0 are treated as ratios (0.5 = 50%).
	// Values > 1.0 are treated as fixed cell counts.
	Size float64
}

// Split divides space between children along an axis.
type Split struct {
	Dir    Direction
	items  []SplitChild
	bounds render.Rect
	rects  []render.Rect
}

// NewHSplit creates a horizontal split (left to right).
func NewHSplit(children ...SplitChild) *Split {
	return &Split{Dir: Horizontal, items: children}
}

// NewVSplit creates a vertical split (top to bottom).
func NewVSplit(children ...SplitChild) *Split {
	return &Split{Dir: Vertical, items: children}
}

// Layout assigns each child its rect within the given bounds.
func (s *Split) Layout(bounds render.Rect) {
	s.bounds = bounds
	s.rects = make([]render.Rect, len(s.items))

	total := s.mainAxis(bounds)
	if total <= 0 || len(s.items) == 0 {
		return
	}

	// First pass: allocate fixed-size children and compute ratio remainder.
	fixedUsed := 0
	ratioSum := 0.0
	for _, c := range s.items {
		if c.Size > 1.0 {
			fixedUsed += int(c.Size)
		} else {
			ratioSum += c.Size
		}
	}

	remaining := max(total-fixedUsed, 0)

	// Second pass: assign rects.
	offset := 0
	for i, c := range s.items {
		var size int
		if c.Size > 1.0 {
			size = int(c.Size)
		} else if ratioSum > 0 {
			size = int(float64(remaining) * c.Size / ratioSum)
		}

		// Last child gets the remainder to avoid rounding gaps.
		if i == len(s.items)-1 {
			size = total - offset
		}
		size = max(size, 0)

		if s.Dir == Horizontal {
			s.rects[i] = render.Rect{
				X:      bounds.X + offset,
				Y:      bounds.Y,
				Width:  size,
				Height: bounds.Height,
			}
		} else {
			s.rects[i] = render.Rect{
				X:      bounds.X,
				Y:      bounds.Y + offset,
				Width:  bounds.Width,
				Height: size,
			}
		}
		offset += size
	}

	// Recursively layout children that are containers.
	for i, c := range s.items {
		if l, ok := c.Renderable.(interface{ Layout(render.Rect) }); ok {
			l.Layout(s.rects[i])
		}
	}
}

// Render renders all children into their assigned regions.
func (s *Split) Render(v *render.View) {
	for i, c := range s.items {
		if s.rects[i].IsEmpty() {
			continue
		}
		sub := v.Sub(s.rects[i].Translate(-s.bounds.X, -s.bounds.Y))
		c.Renderable.Render(sub)
	}
}

// HandleEvent dispatches the event to each child that implements Interactive.
// Returns true if any child consumed it.
func (s *Split) HandleEvent(ev input.Event) bool {
	for _, c := range s.items {
		if h, ok := c.Renderable.(render.Interactive); ok {
			if h.HandleEvent(ev) {
				return true
			}
		}
	}
	return false
}

// Children returns the renderable of each child.
func (s *Split) Children() []render.Renderable {
	rs := make([]render.Renderable, len(s.items))
	for i, c := range s.items {
		rs[i] = c.Renderable
	}
	return rs
}

// ChildBounds returns the bounds assigned to each child after Layout.
func (s *Split) ChildBounds() []render.Rect {
	return s.rects
}

func (s *Split) mainAxis(r render.Rect) int {
	if s.Dir == Horizontal {
		return r.Width
	}
	return r.Height
}
