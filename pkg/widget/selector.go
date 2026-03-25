package widget

import (
	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// Selector is a horizontal option selector widget.
// It implements render.Renderable, render.Interactive, and render.Focusable.
type Selector struct {
	Items    []string
	Selected int
	OnChange func(index int)

	focused bool
	// hitRegions stores the x-range [start, end) for each item, recorded during Render.
	hitRegions []hitRegion
}

type hitRegion struct {
	startX, endX int
}

// NewSelector creates a Selector with the given items.
func NewSelector(items []string) *Selector {
	return &Selector{Items: items}
}

// Focus gives the selector keyboard focus.
func (s *Selector) Focus() { s.focused = true }

// Blur removes keyboard focus from the selector.
func (s *Selector) Blur() { s.focused = false }

// IsFocused reports whether the selector has keyboard focus.
func (s *Selector) IsFocused() bool { return s.focused }

// Render draws the selector into the view.
func (s *Selector) Render(v *render.View) {
	if v.Height() < 1 || len(s.Items) == 0 {
		return
	}

	normalStyle := text.Style{}
	selectedStyle := text.Style{}.Reverse()
	if s.focused {
		selectedStyle = selectedStyle.Bold()
	}
	sepStyle := text.Style{}.Dim()

	s.hitRegions = make([]hitRegion, len(s.Items))
	col := 0
	for i, item := range s.Items {
		if i > 0 {
			col += v.WriteString(col, 0, " | ", sepStyle)
		}
		s.hitRegions[i].startX = col
		st := normalStyle
		if i == s.Selected {
			st = selectedStyle
		}
		col += v.WriteString(col, 0, item, st)
		s.hitRegions[i].endX = col
	}
}

// HandleEvent processes keyboard and mouse events.
func (s *Selector) HandleEvent(ev input.Event) bool {
	switch ev.Type {
	case input.EventKey:
		switch ev.Key.Code {
		case input.KeyLeft:
			if s.Selected > 0 {
				s.Selected--
				s.fireChange()
			}
			return true
		case input.KeyRight:
			if s.Selected < len(s.Items)-1 {
				s.Selected++
				s.fireChange()
			}
			return true
		}
	case input.EventMouse:
		if ev.Mouse.Button == input.MouseLeft {
			mx := ev.Mouse.X
			for i, hr := range s.hitRegions {
				if mx >= hr.startX && mx < hr.endX {
					if s.Selected != i {
						s.Selected = i
						s.fireChange()
					}
					return true
				}
			}
		}
	}
	return false
}

func (s *Selector) fireChange() {
	if s.OnChange != nil {
		s.OnChange(s.Selected)
	}
}
