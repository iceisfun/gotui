package render

import "github.com/iceisfun/gotui/pkg/text"

// Buffer is a 2D grid of cells (the framebuffer).
// New buffers have all cells set to transparent.
type Buffer struct {
	cells  []Cell
	width  int
	height int
}

// NewBuffer creates a buffer of the given dimensions with all cells transparent.
func NewBuffer(w, h int) *Buffer {
	b := &Buffer{
		cells:  make([]Cell, w*h),
		width:  w,
		height: h,
	}
	b.ClearTransparent()
	return b
}

// Width returns the buffer width in columns.
func (b *Buffer) Width() int { return b.width }

// Height returns the buffer height in rows.
func (b *Buffer) Height() int { return b.height }

// Resize changes the buffer dimensions, discarding old content.
func (b *Buffer) Resize(w, h int) {
	if w == b.width && h == b.height {
		return
	}
	b.width = w
	b.height = h
	b.cells = make([]Cell, w*h)
	b.ClearTransparent()
}

// inBounds reports whether (x, y) is within the buffer.
func (b *Buffer) inBounds(x, y int) bool {
	return x >= 0 && x < b.width && y >= 0 && y < b.height
}

// Cell returns the cell at (x, y). Out-of-bounds returns TransparentCell.
func (b *Buffer) Cell(x, y int) Cell {
	if !b.inBounds(x, y) {
		return TransparentCell
	}
	return b.cells[y*b.width+x]
}

// SetCell sets the cell at (x, y). Out-of-bounds writes are silently clipped.
func (b *Buffer) SetCell(x, y int, c Cell) {
	if !b.inBounds(x, y) {
		return
	}
	b.cells[y*b.width+x] = c
}

// SetRune sets an opaque cell at (x, y) with the given rune and style.
func (b *Buffer) SetRune(x, y int, r rune, s text.Style) {
	if !b.inBounds(x, y) {
		return
	}
	b.cells[y*b.width+x] = Cell{Rune: r, Width: 1, Style: s}
}

// WriteString writes a string starting at (x, y) with the given style.
// Returns the number of columns consumed. Wide characters (e.g. CJK)
// occupy two cells: the primary cell holds the rune with Width=2, and
// the next cell is a zero-width continuation cell.
func (b *Buffer) WriteString(x, y int, str string, s text.Style) int {
	col := 0
	for _, r := range str {
		rw := text.RuneWidth(r)
		if rw == 0 {
			continue
		}
		// Stop if the character (including its full width) won't fit.
		if !b.inBounds(x+col, y) {
			break
		}
		if rw == 2 && !b.inBounds(x+col+1, y) {
			break
		}
		b.cells[y*b.width+x+col] = Cell{Rune: r, Width: uint8(rw), Style: s}
		if rw == 2 {
			b.cells[y*b.width+x+col+1] = Cell{Rune: 0, Width: 0, Style: s}
		}
		col += rw
	}
	return col
}

// Fill fills a rectangular region with the given cell.
func (b *Buffer) Fill(r Rect, c Cell) {
	clip := r.Intersect(Rect{X: 0, Y: 0, Width: b.width, Height: b.height})
	for y := clip.Y; y < clip.Y+clip.Height; y++ {
		for x := clip.X; x < clip.X+clip.Width; x++ {
			b.cells[y*b.width+x] = c
		}
	}
}

// Clear fills the entire buffer with opaque blank cells (space, default style).
func (b *Buffer) Clear() {
	for i := range b.cells {
		b.cells[i] = BlankCell
	}
}

// ClearTransparent fills the entire buffer with transparent cells.
func (b *Buffer) ClearTransparent() {
	for i := range b.cells {
		b.cells[i] = TransparentCell
	}
}

// ClearWithStyle fills the entire buffer with opaque space cells in the given style.
func (b *Buffer) ClearWithStyle(s text.Style) {
	c := Cell{Rune: ' ', Width: 1, Style: s}
	for i := range b.cells {
		b.cells[i] = c
	}
}
