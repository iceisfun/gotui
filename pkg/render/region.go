package render

// Rect defines a rectangular region.
type Rect struct {
	X, Y, Width, Height int
}

// Contains reports whether the point (px, py) is inside the rectangle.
func (r Rect) Contains(px, py int) bool {
	return px >= r.X && px < r.X+r.Width && py >= r.Y && py < r.Y+r.Height
}

// Intersect returns the intersection of two rectangles.
// If they do not overlap, the result has zero area.
func (r Rect) Intersect(other Rect) Rect {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.X+r.Width, other.X+other.Width)
	y2 := min(r.Y+r.Height, other.Y+other.Height)
	if x2 <= x1 || y2 <= y1 {
		return Rect{}
	}
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// IsEmpty reports whether the rectangle has zero area.
func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Translate returns a new Rect offset by (dx, dy).
func (r Rect) Translate(dx, dy int) Rect {
	return Rect{X: r.X + dx, Y: r.Y + dy, Width: r.Width, Height: r.Height}
}
