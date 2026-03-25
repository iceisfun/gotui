package widget

import (
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/text"
)

// BorderStyle selects the border drawing characters.
type BorderStyle int

const (
	BorderSingle BorderStyle = iota
	BorderDouble
	BorderRounded
)

// Panel is a bordered container with an optional title.
// It implements render.Renderable and render.Container.
type Panel struct {
	Title   string
	Child   render.Renderable
	Border  BorderStyle
	Focused bool

	childBounds render.Rect
}

// NewPanel creates a Panel wrapping the given child renderable.
func NewPanel(title string, child render.Renderable) *Panel {
	return &Panel{Title: title, Child: child}
}

// Children returns the child renderables.
func (p *Panel) Children() []render.Renderable {
	if p.Child == nil {
		return nil
	}
	return []render.Renderable{p.Child}
}

// Layout assigns bounds to the child (inset by the border).
func (p *Panel) Layout(bounds render.Rect) {
	p.childBounds = render.Rect{
		X:      bounds.X + 1,
		Y:      bounds.Y + 1,
		Width:  bounds.Width - 2,
		Height: bounds.Height - 2,
	}
	if p.childBounds.Width < 0 {
		p.childBounds.Width = 0
	}
	if p.childBounds.Height < 0 {
		p.childBounds.Height = 0
	}
}

// ChildBounds returns the bounds assigned to each child after Layout.
func (p *Panel) ChildBounds() []render.Rect {
	if p.Child == nil {
		return nil
	}
	return []render.Rect{p.childBounds}
}

type borderChars struct {
	topLeft, topRight, bottomLeft, bottomRight rune
	horizontal, vertical                       rune
}

func getBorderChars(bs BorderStyle) borderChars {
	switch bs {
	case BorderDouble:
		return borderChars{
			topLeft: '\u2554', topRight: '\u2557',
			bottomLeft: '\u255A', bottomRight: '\u255D',
			horizontal: '\u2550', vertical: '\u2551',
		}
	case BorderRounded:
		return borderChars{
			topLeft: '\u256D', topRight: '\u256E',
			bottomLeft: '\u2570', bottomRight: '\u256F',
			horizontal: '\u2500', vertical: '\u2502',
		}
	default: // BorderSingle
		return borderChars{
			topLeft: '\u250C', topRight: '\u2510',
			bottomLeft: '\u2514', bottomRight: '\u2518',
			horizontal: '\u2500', vertical: '\u2502',
		}
	}
}

// Render draws the border, title, and child into the view.
func (p *Panel) Render(v *render.View) {
	w, h := v.Width(), v.Height()
	if w < 2 || h < 2 {
		return
	}

	bc := getBorderChars(p.Border)
	borderStyle := text.Style{}
	if p.Focused {
		borderStyle = borderStyle.WithFg(text.Cyan()).Bold()
	}

	// Top border.
	v.SetRune(0, 0, bc.topLeft, borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, 0, bc.horizontal, borderStyle)
	}
	v.SetRune(w-1, 0, bc.topRight, borderStyle)

	// Title.
	if p.Title != "" {
		titleRunes := []rune(p.Title)
		maxTitleLen := w - 4 // borders + spaces
		if maxTitleLen > 0 {
			if len(titleRunes) > maxTitleLen {
				titleRunes = titleRunes[:maxTitleLen]
			}
			titleStyle := borderStyle
			v.SetRune(1, 0, ' ', borderStyle)
			v.WriteString(2, 0, string(titleRunes), titleStyle)
			v.SetRune(2+len(titleRunes), 0, ' ', borderStyle)
		}
	}

	// Side borders.
	for y := 1; y < h-1; y++ {
		v.SetRune(0, y, bc.vertical, borderStyle)
		v.SetRune(w-1, y, bc.vertical, borderStyle)
	}

	// Bottom border.
	v.SetRune(0, h-1, bc.bottomLeft, borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, h-1, bc.horizontal, borderStyle)
	}
	v.SetRune(w-1, h-1, bc.bottomRight, borderStyle)

	// Render child into sub-view.
	if p.Child != nil && w > 2 && h > 2 {
		childView := v.Sub(render.Rect{X: 1, Y: 1, Width: w - 2, Height: h - 2})
		p.Child.Render(childView)
	}
}
