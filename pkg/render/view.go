package render

import "github.com/iceisfun/gotui/pkg/text"

// View provides a clipped, offset window into a Buffer.
// All coordinates are relative to the view's origin. Writes outside the
// view bounds are silently clipped.
type View struct {
	buf    *Buffer
	bounds Rect
}

// NewView creates a View into buf clipped to bounds.
func NewView(buf *Buffer, bounds Rect) *View {
	return &View{buf: buf, bounds: bounds}
}

// Width returns the view width.
func (v *View) Width() int { return v.bounds.Width }

// Height returns the view height.
func (v *View) Height() int { return v.bounds.Height }

// Bounds returns the view's rectangle in buffer coordinates.
func (v *View) Bounds() Rect { return v.bounds }

// SetCell sets a cell at view-relative (x, y).
func (v *View) SetCell(x, y int, c Cell) {
	bx, by := v.bounds.X+x, v.bounds.Y+y
	if !v.contains(bx, by) {
		return
	}
	v.buf.SetCell(bx, by, c)
}

// SetRune sets an opaque cell at view-relative (x, y).
func (v *View) SetRune(x, y int, r rune, s text.Style) {
	bx, by := v.bounds.X+x, v.bounds.Y+y
	if !v.contains(bx, by) {
		return
	}
	v.buf.SetRune(bx, by, r, s)
}

// WriteString writes a string at view-relative (x, y), clipping at the view edge.
// Returns columns consumed. Wide characters occupy two cells.
func (v *View) WriteString(x, y int, str string, s text.Style) int {
	by := v.bounds.Y + y
	if by < v.bounds.Y || by >= v.bounds.Y+v.bounds.Height {
		return 0
	}

	col := 0
	for _, r := range str {
		rw := text.RuneWidth(r)
		if rw == 0 {
			continue
		}
		bx := v.bounds.X + x + col
		if bx < v.bounds.X {
			col += rw
			continue
		}
		rightEdge := v.bounds.X + v.bounds.Width
		if bx >= rightEdge {
			break
		}
		if rw == 2 && bx+1 >= rightEdge {
			break
		}
		v.buf.SetCell(bx, by, Cell{Rune: r, Width: uint8(rw), Style: s})
		if rw == 2 {
			v.buf.SetCell(bx+1, by, Cell{Rune: 0, Width: 0, Style: s})
		}
		col += rw
	}
	return col
}

// Fill fills a view-relative rectangle with the given cell.
func (v *View) Fill(r Rect, c Cell) {
	// Translate to buffer coordinates and clip.
	br := Rect{
		X:      v.bounds.X + r.X,
		Y:      v.bounds.Y + r.Y,
		Width:  r.Width,
		Height: r.Height,
	}
	clip := br.Intersect(v.bounds)
	v.buf.Fill(clip, c)
}

// Clear fills the entire view with opaque blank cells.
func (v *View) Clear() {
	v.buf.Fill(v.bounds, BlankCell)
}

// ClearTransparent fills the entire view with transparent cells.
func (v *View) ClearTransparent() {
	v.buf.Fill(v.bounds, TransparentCell)
}

// Sub creates a sub-view within this view. The rect is relative to this view.
func (v *View) Sub(r Rect) *View {
	abs := Rect{
		X:      v.bounds.X + r.X,
		Y:      v.bounds.Y + r.Y,
		Width:  r.Width,
		Height: r.Height,
	}
	clipped := abs.Intersect(v.bounds)
	return NewView(v.buf, clipped)
}

func (v *View) contains(bx, by int) bool {
	return bx >= v.bounds.X && bx < v.bounds.X+v.bounds.Width &&
		by >= v.bounds.Y && by < v.bounds.Y+v.bounds.Height
}
