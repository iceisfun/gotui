package repl

import (
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// Output is a scrollable area that displays REPL output (results and errors).
type Output struct {
	lines   []text.StyledLine
	scrollY int
}

// NewOutput creates an empty output area.
func NewOutput() *Output {
	return &Output{}
}

// Append adds styled lines to the output.
func (o *Output) Append(lines ...text.StyledLine) {
	o.lines = append(o.lines, lines...)
}

// AppendText adds a plain text string to the output.
func (o *Output) AppendText(s string, style text.Style) {
	o.lines = append(o.lines, text.Styled(s, style))
}

// Clear removes all output.
func (o *Output) Clear() {
	o.lines = nil
	o.scrollY = 0
}

// Lines returns the total number of output lines.
func (o *Output) Lines() int {
	return len(o.lines)
}

// Render draws the output into the given view, respecting the scroll offset.
// Shows the tail (most recent) by default.
func (o *Output) Render(v *render.View, startY int) int {
	h := v.Height() - startY
	if h <= 0 || len(o.lines) == 0 {
		return 0
	}

	// Show the last lines that fit, adjusted by scrollY.
	totalVisible := min(len(o.lines), h)
	end := len(o.lines) - o.scrollY
	if end < totalVisible {
		end = totalVisible
	}
	if end > len(o.lines) {
		end = len(o.lines)
	}
	start := end - totalVisible
	if start < 0 {
		start = 0
	}

	rendered := 0
	for i := start; i < end && startY+rendered < v.Height(); i++ {
		line := o.lines[i]
		col := 0
		for _, span := range line {
			col += v.WriteString(col, startY+rendered, span.Text, span.Style)
		}
		rendered++
	}
	return rendered
}

// ScrollUp scrolls up by n lines.
func (o *Output) ScrollUp(n int) {
	o.scrollY = min(o.scrollY+n, max(len(o.lines)-1, 0))
}

// ScrollDown scrolls down by n lines.
func (o *Output) ScrollDown(n int) {
	o.scrollY = max(o.scrollY-n, 0)
}
