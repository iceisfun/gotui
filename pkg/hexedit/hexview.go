package hexedit

import (
	"fmt"

	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

const bytesPerRow = 16

// HexView is the main hex editor renderable.
type HexView struct {
	buf     *Buffer
	cursor  int // Byte offset of cursor.
	scrollY int // First visible row.
	focused bool

	// Visual dimensions from last render.
	viewWidth  int
	viewHeight int

	// Search highlight.
	highlightStart int
	highlightLen   int

	// Styles.
	offsetStyle    text.Style
	hexStyle       text.Style
	hexAltStyle    text.Style
	asciiStyle     text.Style
	cursorStyle    text.Style
	highlightStyle text.Style
	headerStyle    text.Style
	nonPrintStyle  text.Style
}

// NewHexView creates a hex view for the given buffer.
func NewHexView(buf *Buffer) *HexView {
	return &HexView{
		buf:            buf,
		highlightStart: -1,
		offsetStyle:    text.Style{}.WithFg(text.BrightBlack()),
		hexStyle:       text.Style{}.WithFg(text.White()),
		hexAltStyle:    text.Style{}.WithFg(text.Color256(250)),
		asciiStyle:     text.Style{}.WithFg(text.Green()),
		cursorStyle:    text.Style{}.WithFg(text.Black()).WithBg(text.Cyan()),
		highlightStyle: text.Style{}.WithFg(text.Black()).WithBg(text.Yellow()),
		headerStyle:    text.Style{}.WithFg(text.Cyan()).Bold(),
		nonPrintStyle:  text.Style{}.WithFg(text.BrightBlack()),
	}
}

// Buffer returns the underlying buffer.
func (h *HexView) Buffer() *Buffer { return h.buf }

// Cursor returns the current cursor offset.
func (h *HexView) Cursor() int { return h.cursor }

// SetCursor moves the cursor and adjusts scroll.
func (h *HexView) SetCursor(offset int) {
	if offset < 0 {
		offset = 0
	}
	if offset >= h.buf.Len() {
		offset = max(h.buf.Len()-1, 0)
	}
	h.cursor = offset
	h.ensureVisible()
}

// SetHighlight sets a highlighted byte range (for search results).
func (h *HexView) SetHighlight(start, length int) {
	h.highlightStart = start
	h.highlightLen = length
}

// ClearHighlight removes the search highlight.
func (h *HexView) ClearHighlight() {
	h.highlightStart = -1
	h.highlightLen = 0
}

// GotoOffset jumps to an absolute byte offset.
func (h *HexView) GotoOffset(offset int) {
	h.SetCursor(offset)
}

// TotalRows returns the total number of rows in the file.
func (h *HexView) TotalRows() int {
	return (h.buf.Len() + bytesPerRow - 1) / bytesPerRow
}

// VisibleRows returns the number of data rows visible (excluding header).
func (h *HexView) VisibleRows() int {
	return max(h.viewHeight-1, 1) // -1 for header row.
}

// Render draws the hex view: header + offset | hex bytes | ascii.
func (h *HexView) Render(v *render.View) {
	h.viewWidth = v.Width()
	h.viewHeight = v.Height()

	if h.buf.Len() == 0 {
		v.WriteString(1, 0, "(empty file)", text.Style{}.WithFg(text.BrightBlack()))
		return
	}

	// Header row.
	v.WriteString(0, 0, "  Offset  ", h.headerStyle)
	for i := range bytesPerRow {
		x := 10 + i*3
		v.WriteString(x, 0, fmt.Sprintf("%02X", i), h.headerStyle)
		if i == 7 {
			v.SetRune(x+2, 0, ' ', h.headerStyle)
		}
	}
	asciiX := 10 + bytesPerRow*3 + 1
	v.WriteString(asciiX, 0, "ASCII", h.headerStyle)

	// Data rows.
	visibleRows := h.VisibleRows()
	for row := range visibleRows {
		fileRow := h.scrollY + row
		offset := fileRow * bytesPerRow
		if offset >= h.buf.Len() {
			break
		}

		y := row + 1 // +1 for header.

		// Offset column.
		v.WriteString(0, y, fmt.Sprintf("%08X  ", offset), h.offsetStyle)

		// Hex bytes.
		for col := range bytesPerRow {
			byteOffset := offset + col
			x := 10 + col*3
			if col >= 8 {
				x++ // Extra space between byte groups.
			}

			if byteOffset >= h.buf.Len() {
				v.WriteString(x, y, "  ", h.hexStyle)
				continue
			}

			b := h.buf.Byte(byteOffset)
			hexStr := fmt.Sprintf("%02X", b)

			style := h.hexStyle
			if col%2 == 1 {
				style = h.hexAltStyle
			}
			if h.isHighlighted(byteOffset) {
				style = h.highlightStyle
			}
			if byteOffset == h.cursor && h.focused {
				style = h.cursorStyle
			}

			v.WriteString(x, y, hexStr, style)
		}

		// ASCII column.
		asciiStartX := 10 + bytesPerRow*3 + 2
		for col := range bytesPerRow {
			byteOffset := offset + col
			if byteOffset >= h.buf.Len() {
				break
			}

			b := h.buf.Byte(byteOffset)
			r := rune(b)
			style := h.asciiStyle
			if b < 0x20 || b > 0x7e {
				r = '.'
				style = h.nonPrintStyle
			}
			if h.isHighlighted(byteOffset) {
				style = h.highlightStyle
			}
			if byteOffset == h.cursor && h.focused {
				style = h.cursorStyle
			}

			v.SetRune(asciiStartX+col, y, r, style)
		}
	}
}

// HandleEvent processes navigation keys.
func (h *HexView) HandleEvent(ev input.Event) bool {
	if ev.Type != input.EventKey {
		return false
	}

	key := ev.Key
	switch key.Code {
	case input.KeyLeft:
		h.SetCursor(h.cursor - 1)
		return true
	case input.KeyRight:
		h.SetCursor(h.cursor + 1)
		return true
	case input.KeyUp:
		h.SetCursor(h.cursor - bytesPerRow)
		return true
	case input.KeyDown:
		h.SetCursor(h.cursor + bytesPerRow)
		return true
	case input.KeyHome:
		if key.Mod&input.ModCtrl != 0 {
			h.SetCursor(0)
		} else {
			h.SetCursor(h.cursor - h.cursor%bytesPerRow)
		}
		return true
	case input.KeyEnd:
		if key.Mod&input.ModCtrl != 0 {
			h.SetCursor(h.buf.Len() - 1)
		} else {
			h.SetCursor(h.cursor - h.cursor%bytesPerRow + bytesPerRow - 1)
		}
		return true
	case input.KeyPgUp:
		h.SetCursor(h.cursor - h.VisibleRows()*bytesPerRow)
		return true
	case input.KeyPgDown:
		h.SetCursor(h.cursor + h.VisibleRows()*bytesPerRow)
		return true
	}

	return false
}

func (h *HexView) Focus()          { h.focused = true }
func (h *HexView) Blur()           { h.focused = false }
func (h *HexView) IsFocused() bool { return h.focused }

func (h *HexView) CursorPosition() (x, y int, visible bool) {
	if !h.focused {
		return 0, 0, false
	}
	row := h.cursor/bytesPerRow - h.scrollY
	col := h.cursor % bytesPerRow
	x = 10 + col*3
	if col >= 8 {
		x++
	}
	y = row + 1 // +1 for header.
	return x, y, y >= 1 && y < h.viewHeight
}

func (h *HexView) ensureVisible() {
	row := h.cursor / bytesPerRow
	visibleRows := h.VisibleRows()
	if row < h.scrollY {
		h.scrollY = row
	} else if row >= h.scrollY+visibleRows {
		h.scrollY = row - visibleRows + 1
	}
}

func (h *HexView) isHighlighted(offset int) bool {
	if h.highlightStart < 0 {
		return false
	}
	return offset >= h.highlightStart && offset < h.highlightStart+h.highlightLen
}
