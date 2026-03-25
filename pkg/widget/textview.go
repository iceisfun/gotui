package widget

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/text"
)

// TextView is a scrollable multi-line text display widget.
// It implements render.Renderable, render.Interactive, and render.Scrollable.
type TextView struct {
	lines      []text.StyledLine
	scrollY    int
	AutoScroll bool
}

// NewTextView creates an empty TextView with auto-scroll enabled.
func NewTextView() *TextView {
	return &TextView{AutoScroll: true}
}

// SetLines replaces all content.
func (tv *TextView) SetLines(lines []text.StyledLine) {
	tv.lines = lines
	if tv.AutoScroll {
		tv.scrollY = len(tv.lines) // will be clamped on render
	}
}

// AppendLine adds a line to the end of the content.
func (tv *TextView) AppendLine(line text.StyledLine) {
	tv.lines = append(tv.lines, line)
	if tv.AutoScroll {
		tv.scrollY = len(tv.lines) // will be clamped on render
	}
}

// AppendPlain adds a plain text line.
func (tv *TextView) AppendPlain(s string) {
	tv.AppendLine(text.Plain(s))
}

// LineCount returns the number of lines.
func (tv *TextView) LineCount() int { return len(tv.lines) }

// ContentSize returns the total content dimensions.
func (tv *TextView) ContentSize() (w, h int) {
	maxW := 0
	for _, line := range tv.lines {
		if lw := line.Len(); lw > maxW {
			maxW = lw
		}
	}
	return maxW, len(tv.lines)
}

// ScrollOffset returns the current scroll position.
func (tv *TextView) ScrollOffset() (x, y int) { return 0, tv.scrollY }

// SetScrollOffset sets the scroll position.
func (tv *TextView) SetScrollOffset(x, y int) {
	tv.scrollY = y
	if tv.scrollY < 0 {
		tv.scrollY = 0
	}
	tv.AutoScroll = false
}

// Render draws the text into the view.
func (tv *TextView) Render(v *render.View) {
	h := v.Height()
	if h < 1 {
		return
	}

	tv.clampScroll(h)

	for row := 0; row < h; row++ {
		idx := row + tv.scrollY
		if idx >= len(tv.lines) {
			break
		}
		col := 0
		for _, span := range tv.lines[idx] {
			col += v.WriteString(col, row, span.Text, span.Style)
		}
	}
}

// HandleEvent processes mouse wheel events for scrolling.
func (tv *TextView) HandleEvent(ev input.Event) bool {
	if ev.Type != input.EventMouse {
		return false
	}
	switch ev.Mouse.Button {
	case input.MouseWheelUp:
		tv.scrollY -= 3
		if tv.scrollY < 0 {
			tv.scrollY = 0
		}
		tv.AutoScroll = false
		return true
	case input.MouseWheelDown:
		tv.scrollY += 3
		tv.AutoScroll = false
		return true
	}
	return false
}

func (tv *TextView) clampScroll(viewHeight int) {
	maxScroll := len(tv.lines) - viewHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if tv.scrollY > maxScroll {
		tv.scrollY = maxScroll
	}
	if tv.scrollY < 0 {
		tv.scrollY = 0
	}
}
