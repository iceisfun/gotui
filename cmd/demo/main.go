package main

import (
	"context"
	"fmt"
	"os"

	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/layout"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/term"
	"github.com/iceisfun/gotui/pkg/text"
)

// ClickPanel is a panel with clickable buttons inside it.
// It tests that local mouse coordinates are correct through the layout tree.
type ClickPanel struct {
	name    string
	focused bool
	log     []text.StyledLine
	buttons []button
	maxLog  int
}

type button struct {
	label string
	x, y  int
	w     int
}

func NewClickPanel(name string) *ClickPanel {
	p := &ClickPanel{
		name:   name,
		maxLog: 50,
		buttons: []button{
			{label: "[Red]", x: 2, y: 2, w: 5},
			{label: "[Green]", x: 9, y: 2, w: 7},
			{label: "[Blue]", x: 18, y: 2, w: 6},
			{label: "[Reset]", x: 26, y: 2, w: 7},
		},
	}
	p.addLog("Click the buttons or anywhere inside")
	return p
}

func (p *ClickPanel) addLog(msg string) {
	p.log = append(p.log, text.Parse(msg))
	if len(p.log) > p.maxLog {
		p.log = p.log[len(p.log)-p.maxLog:]
	}
}

func (p *ClickPanel) Render(v *render.View) {
	w, h := v.Width(), v.Height()
	if w < 4 || h < 4 {
		return
	}

	borderStyle := text.Style{}.WithFg(text.BrightBlack())
	titleStyle := text.Style{}.WithFg(text.White())
	if p.focused {
		borderStyle = text.Style{}.WithFg(text.Cyan()).Bold()
		titleStyle = text.Style{}.WithFg(text.Cyan()).Bold()
	}

	// Border.
	v.SetRune(0, 0, '┌', borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, 0, '─', borderStyle)
	}
	v.SetRune(w-1, 0, '┐', borderStyle)
	title := " " + p.name + " "
	if len(title)+2 < w {
		v.WriteString(2, 0, title, titleStyle)
	}
	for y := 1; y < h-1; y++ {
		v.SetRune(0, y, '│', borderStyle)
		v.SetRune(w-1, y, '│', borderStyle)
	}
	v.SetRune(0, h-1, '└', borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, h-1, '─', borderStyle)
	}
	v.SetRune(w-1, h-1, '┘', borderStyle)

	inner := v.Sub(render.Rect{X: 1, Y: 1, Width: w - 2, Height: h - 2})

	// Instruction.
	inner.WriteString(1, 0, "Buttons (click them):", text.Style{}.WithFg(text.BrightBlack()))

	// Draw buttons.
	btnStyles := map[string]text.Style{
		"[Red]":   text.Style{}.WithFg(text.White()).WithBg(text.Red()),
		"[Green]": text.Style{}.WithFg(text.Black()).WithBg(text.Green()),
		"[Blue]":  text.Style{}.WithFg(text.White()).WithBg(text.Blue()),
		"[Reset]": text.Style{}.WithFg(text.Black()).WithBg(text.White()),
	}
	for _, btn := range p.buttons {
		style := btnStyles[btn.label]
		inner.WriteString(btn.x, btn.y, btn.label, style)
	}

	// Log area.
	logStart := 4
	logH := inner.Height() - logStart
	if logH <= 0 {
		return
	}
	start := 0
	if len(p.log) > logH {
		start = len(p.log) - logH
	}
	for i, line := range p.log[start:] {
		y := logStart + i
		col := 0
		for _, span := range line {
			col += inner.WriteString(col, y, span.Text, span.Style)
		}
	}
}

func (p *ClickPanel) HandleEvent(ev input.Event) bool {
	if ev.Type == input.EventMouse {
		m := ev.Mouse
		if m.Button == input.MouseLeft {
			// Coords are local to this renderable. Adjust for border.
			lx, ly := m.X-1, m.Y-1

			// Check button hits.
			for _, btn := range p.buttons {
				if ly == btn.y && lx >= btn.x && lx < btn.x+btn.w {
					p.addLog(fmt.Sprintf("[bold]Button %s[/] clicked! local=(%d,%d) screen=(%d,%d)",
						btn.label, m.X, m.Y, m.ScreenX, m.ScreenY))
					return true
				}
			}

			// Click elsewhere in panel.
			p.addLog(fmt.Sprintf("[dim]Panel click[/] local=(%d,%d) screen=(%d,%d)",
				m.X, m.Y, m.ScreenX, m.ScreenY))
			return true
		}
	}

	if ev.Type == input.EventKey {
		p.addLog(fmt.Sprintf("[green]KEY:[/] %s", ev.Key.String()))
		return true
	}

	return false
}

func (p *ClickPanel) Focus()          { p.focused = true }
func (p *ClickPanel) Blur()           { p.focused = false }
func (p *ClickPanel) IsFocused() bool { return p.focused }

// OverlayPanel is a floating panel that tests overlay hit-testing.
type OverlayPanel struct {
	log     []string
	buttons []button
}

func NewOverlayPanel() *OverlayPanel {
	return &OverlayPanel{
		buttons: []button{
			{label: "[OK]", x: 2, y: 4, w: 4},
			{label: "[Cancel]", x: 8, y: 4, w: 8},
		},
	}
}

func (o *OverlayPanel) Render(v *render.View) {
	w, h := v.Width(), v.Height()
	bgStyle := text.Style{}.WithBg(text.Color256(235))
	borderStyle := text.Style{}.WithFg(text.Yellow())

	// Fill background.
	for y := range h {
		for x := range w {
			v.SetRune(x, y, ' ', bgStyle)
		}
	}

	// Border.
	v.SetRune(0, 0, '┌', borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, 0, '─', borderStyle)
	}
	v.SetRune(w-1, 0, '┐', borderStyle)
	v.WriteString(2, 0, " Overlay ", text.Style{}.WithFg(text.Yellow()).Bold())
	for y := 1; y < h-1; y++ {
		v.SetRune(0, y, '│', borderStyle)
		v.SetRune(w-1, y, '│', borderStyle)
	}
	v.SetRune(0, h-1, '└', borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, h-1, '─', borderStyle)
	}
	v.SetRune(w-1, h-1, '┘', borderStyle)

	v.WriteString(2, 1, "This is a floating overlay.", bgStyle)
	v.WriteString(2, 2, "Click buttons below.", bgStyle)
	v.WriteString(2, 3, "Press Esc to close.", bgStyle)

	okStyle := text.Style{}.WithFg(text.Black()).WithBg(text.Green())
	cancelStyle := text.Style{}.WithFg(text.Black()).WithBg(text.Red())
	v.WriteString(2, 4, "[OK]", okStyle)
	v.WriteString(8, 4, "[Cancel]", cancelStyle)

	// Show log.
	for i, msg := range o.log {
		if 6+i >= h-1 {
			break
		}
		v.WriteString(2, 6+i, msg, text.Style{}.WithFg(text.White()))
	}
}

func (o *OverlayPanel) HandleEvent(ev input.Event) bool {
	if ev.Type == input.EventMouse && ev.Mouse.Button == input.MouseLeft {
		mx, my := ev.Mouse.X, ev.Mouse.Y
		for _, btn := range o.buttons {
			if my == btn.y && mx >= btn.x && mx < btn.x+btn.w {
				msg := fmt.Sprintf("Clicked %s at local=(%d,%d) screen=(%d,%d)",
					btn.label, mx, my, ev.Mouse.ScreenX, ev.Mouse.ScreenY)
				o.log = append(o.log, msg)
				if len(o.log) > 3 {
					o.log = o.log[len(o.log)-3:]
				}
				return true
			}
		}
		return true // Consume clicks inside overlay.
	}
	return false
}

// OverlayHost wraps a child and produces an overlay when toggled.
type OverlayHost struct {
	child       render.Renderable
	overlay     *OverlayPanel
	showOverlay bool
	bounds      render.Rect
}

func (h *OverlayHost) Render(v *render.View)      { h.child.Render(v) }
func (h *OverlayHost) Layout(bounds render.Rect) {
	h.bounds = bounds
	if l, ok := h.child.(interface{ Layout(render.Rect) }); ok {
		l.Layout(bounds)
	}
}
func (h *OverlayHost) Children() []render.Renderable { return []render.Renderable{h.child} }
func (h *OverlayHost) ChildBounds() []render.Rect    { return []render.Rect{h.bounds} }

func (h *OverlayHost) HandleEvent(ev input.Event) bool {
	if i, ok := h.child.(render.Interactive); ok {
		return i.HandleEvent(ev)
	}
	return false
}

func (h *OverlayHost) Overlays() []render.OverlayRequest {
	if !h.showOverlay {
		return nil
	}
	return []render.OverlayRequest{{
		Renderable: h.overlay,
		Anchor:     render.Rect{X: 5, Y: 3, Width: 35, Height: 10},
		ZOrder:     100,
	}}
}

func main() {
	t, err := term.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()

	left := NewClickPanel("Left Panel")
	right := NewClickPanel("Right Panel")

	split := layout.NewHSplit(
		layout.SplitChild{Renderable: left, Size: 0.5},
		layout.SplitChild{Renderable: right, Size: 0.5},
	)

	overlayPanel := NewOverlayPanel()
	host := &OverlayHost{
		child:   split,
		overlay: overlayPanel,
	}

	app := layout.NewApp(t, host)
	app.SetFocus(left)

	// Tab cycles focus.
	focusables := []render.Focusable{left, right}
	focusIdx := 0
	app.BindKey(input.KeyTab, 0, func() {
		focusIdx = (focusIdx + 1) % len(focusables)
		app.SetFocus(focusables[focusIdx])
	})

	// Ctrl-O toggles overlay.
	app.BindRune('o', input.ModCtrl, func() {
		host.showOverlay = !host.showOverlay
		app.RequestRender()
	})

	// Escape closes overlay.
	app.BindKey(input.KeyEscape, 0, func() {
		if host.showOverlay {
			host.showOverlay = false
			app.RequestRender()
		}
	})

	// Ctrl-C quits.
	app.BindRune('c', input.ModCtrl, func() { app.Quit() })

	if err := app.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
