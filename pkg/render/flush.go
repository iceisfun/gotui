package render

import (
	"github.com/iceisfun/gorepl/pkg/term"
	"github.com/iceisfun/gorepl/pkg/text"
)

// Flush writes the differences between prev and curr to the terminal writer.
// Only changed cells emit ANSI sequences. After flushing, prev is updated
// to match curr so the next frame can diff against it.
func Flush(w *term.Writer, prev, curr *Buffer) {
	width, height := curr.Width(), curr.Height()
	var lastStyle text.Style
	styleSet := false
	cursorRow, cursorCol := -1, -1

	w.HideCursor()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			c := curr.cells[idx]
			p := prev.cells[idx]

			// Skip unchanged cells.
			if c.Rune == p.Rune && c.Style.Equal(p.Style) && c.Transparent == p.Transparent && c.Width == p.Width {
				continue
			}

			// Position the cursor if not already there.
			if cursorRow != y || cursorCol != x {
				w.MoveTo(y+1, x+1) // 1-based
				cursorRow = y
				cursorCol = x
			}

			// Update style if changed.
			if !styleSet || !c.Style.Equal(lastStyle) {
				w.SetStyle(c.Style)
				lastStyle = c.Style
				styleSet = true
			}

			// Write the character.
			if c.Transparent || c.Width == 0 {
				w.WriteRune(' ')
			} else {
				w.WriteRune(c.Rune)
			}
			cursorCol++
		}
	}

	// Copy curr into prev for next frame.
	copy(prev.cells, curr.cells)
}
