package widget

import (
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// Alignment controls horizontal text placement.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Label is a static styled text display widget.
// It implements render.Renderable.
type Label struct {
	Line  text.StyledLine
	Align Alignment
}

// NewLabel creates a Label from a plain string.
func NewLabel(s string) *Label {
	return &Label{Line: text.Plain(s)}
}

// NewStyledLabel creates a Label from a StyledLine.
func NewStyledLabel(line text.StyledLine) *Label {
	return &Label{Line: line}
}

// SetText replaces the label text with a plain string.
func (l *Label) SetText(s string) {
	l.Line = text.Plain(s)
}

// SetStyledText replaces the label text with a StyledLine.
func (l *Label) SetStyledText(line text.StyledLine) {
	l.Line = line
}

// Render draws the label into the view.
func (l *Label) Render(v *render.View) {
	if v.Height() < 1 {
		return
	}

	lineLen := l.Line.Len()
	var offsetX int
	switch l.Align {
	case AlignCenter:
		offsetX = (v.Width() - lineLen) / 2
		if offsetX < 0 {
			offsetX = 0
		}
	case AlignRight:
		offsetX = v.Width() - lineLen
		if offsetX < 0 {
			offsetX = 0
		}
	default:
		offsetX = 0
	}

	col := offsetX
	for _, span := range l.Line {
		col += v.WriteString(col, 0, span.Text, span.Style)
	}
}
