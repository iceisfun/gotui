package widget

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/text"
)

// List is a vertical scrollable list widget.
// It implements render.Renderable, render.Interactive, render.Focusable, and render.Scrollable.
type List struct {
	Items    []string
	Selected int
	OnSelect func(index int)

	focused bool
	scrollY int
}

// NewList creates a List with the given items.
func NewList(items []string) *List {
	return &List{Items: items}
}

// Focus gives the list keyboard focus.
func (l *List) Focus() { l.focused = true }

// Blur removes keyboard focus from the list.
func (l *List) Blur() { l.focused = false }

// IsFocused reports whether the list has keyboard focus.
func (l *List) IsFocused() bool { return l.focused }

// ContentSize returns the total content dimensions.
func (l *List) ContentSize() (w, h int) {
	maxW := 0
	for _, item := range l.Items {
		if rw := len([]rune(item)); rw > maxW {
			maxW = rw
		}
	}
	return maxW, len(l.Items)
}

// ScrollOffset returns the current scroll position.
func (l *List) ScrollOffset() (x, y int) { return 0, l.scrollY }

// SetScrollOffset sets the scroll position.
func (l *List) SetScrollOffset(x, y int) {
	l.scrollY = y
	if l.scrollY < 0 {
		l.scrollY = 0
	}
}

// Render draws the list into the view.
func (l *List) Render(v *render.View) {
	h := v.Height()
	if h < 1 || len(l.Items) == 0 {
		return
	}

	l.clampScroll(h)

	normalStyle := text.Style{}
	selectedStyle := text.Style{}.Reverse()
	if l.focused {
		selectedStyle = selectedStyle.Bold()
	}

	for row := 0; row < h; row++ {
		idx := row + l.scrollY
		if idx >= len(l.Items) {
			break
		}
		st := normalStyle
		if idx == l.Selected {
			st = selectedStyle
		}
		v.WriteString(0, row, l.Items[idx], st)
	}
}

// HandleEvent processes keyboard and mouse events.
func (l *List) HandleEvent(ev input.Event) bool {
	switch ev.Type {
	case input.EventKey:
		switch ev.Key.Code {
		case input.KeyUp:
			if l.Selected > 0 {
				l.Selected--
			}
			return true
		case input.KeyDown:
			if l.Selected < len(l.Items)-1 {
				l.Selected++
			}
			return true
		case input.KeyEnter:
			if l.OnSelect != nil && l.Selected >= 0 && l.Selected < len(l.Items) {
				l.OnSelect(l.Selected)
			}
			return true
		case input.KeyPgUp:
			l.Selected -= 10
			if l.Selected < 0 {
				l.Selected = 0
			}
			return true
		case input.KeyPgDown:
			l.Selected += 10
			if l.Selected >= len(l.Items) {
				l.Selected = len(l.Items) - 1
			}
			return true
		case input.KeyHome:
			l.Selected = 0
			return true
		case input.KeyEnd:
			l.Selected = len(l.Items) - 1
			return true
		}
	case input.EventMouse:
		switch ev.Mouse.Button {
		case input.MouseLeft:
			idx := ev.Mouse.Y + l.scrollY
			if idx >= 0 && idx < len(l.Items) {
				l.Selected = idx
				if l.OnSelect != nil {
					l.OnSelect(l.Selected)
				}
			}
			return true
		case input.MouseWheelUp:
			l.scrollY -= 3
			if l.scrollY < 0 {
				l.scrollY = 0
			}
			return true
		case input.MouseWheelDown:
			l.scrollY += 3
			return true
		}
	}
	return false
}

func (l *List) clampScroll(viewHeight int) {
	// Ensure selected item is visible.
	if l.Selected < l.scrollY {
		l.scrollY = l.Selected
	}
	if l.Selected >= l.scrollY+viewHeight {
		l.scrollY = l.Selected - viewHeight + 1
	}
	// Clamp scroll to valid range.
	maxScroll := len(l.Items) - viewHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if l.scrollY > maxScroll {
		l.scrollY = maxScroll
	}
	if l.scrollY < 0 {
		l.scrollY = 0
	}
}
