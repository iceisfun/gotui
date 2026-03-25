package widget

import (
	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// Button is a clickable button widget.
// It implements render.Renderable, render.Interactive, and render.Focusable.
type Button struct {
	Label   string
	OnClick func()

	focused bool
	pressed bool
}

// NewButton creates a Button with the given label and click handler.
func NewButton(label string, onClick func()) *Button {
	return &Button{Label: label, OnClick: onClick}
}

// Focus gives the button keyboard focus.
func (b *Button) Focus() { b.focused = true }

// Blur removes keyboard focus from the button.
func (b *Button) Blur() { b.focused = false; b.pressed = false }

// IsFocused reports whether the button has keyboard focus.
func (b *Button) IsFocused() bool { return b.focused }

// Render draws the button into the view.
func (b *Button) Render(v *render.View) {
	if v.Height() < 1 {
		return
	}

	style := text.Style{}
	switch {
	case b.pressed:
		style = style.Reverse().Bold()
	case b.focused:
		style = style.Bold().WithFg(text.Cyan())
	}

	display := "[ " + b.Label + " ]"
	v.WriteString(0, 0, display, style)
}

// HandleEvent processes keyboard and mouse events.
func (b *Button) HandleEvent(ev input.Event) bool {
	switch ev.Type {
	case input.EventKey:
		if ev.Key.Code == input.KeyEnter {
			b.pressed = true
			if b.OnClick != nil {
				b.OnClick()
			}
			b.pressed = false
			return true
		}
	case input.EventMouse:
		if ev.Mouse.Button == input.MouseLeft {
			b.pressed = true
			if b.OnClick != nil {
				b.OnClick()
			}
			b.pressed = false
			return true
		}
	}
	return false
}
