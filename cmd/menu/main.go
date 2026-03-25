package main

import (
	"context"
	"fmt"
	"os"
	"unicode"

	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/layout"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/term"
	"github.com/iceisfun/gotui/pkg/text"
)

// MenuItem represents one entry in a dropdown menu.
type MenuItem struct {
	Label    string // Display label; empty for separator.
	Shortcut string // Right-aligned shortcut hint.
	IsSep    bool   // Separator line.
}

// Menu represents a top-level menu with its dropdown items.
type Menu struct {
	Title string
	Items []MenuItem
}

// DropdownRenderable renders a bordered dropdown menu.
type DropdownRenderable struct {
	items    []MenuItem
	selected int
	width    int
	height   int
}

func NewDropdownRenderable(items []MenuItem) *DropdownRenderable {
	// Calculate width: border + padding + longest label + shortcut gap.
	maxW := 0
	for _, it := range items {
		w := len([]rune(it.Label))
		if it.Shortcut != "" {
			w += 2 + len([]rune(it.Shortcut)) // gap + shortcut
		}
		if w > maxW {
			maxW = w
		}
	}
	width := maxW + 4 // 2 border + 2 padding
	if width < 12 {
		width = 12
	}
	height := len(items) + 2 // items + top/bottom border
	return &DropdownRenderable{
		items:    items,
		selected: firstSelectable(items),
		width:    width,
		height:   height,
	}
}

func firstSelectable(items []MenuItem) int {
	for i, it := range items {
		if !it.IsSep {
			return i
		}
	}
	return 0
}

func (d *DropdownRenderable) Render(v *render.View) {
	w, h := d.width, d.height
	if w > v.Width() {
		w = v.Width()
	}
	if h > v.Height() {
		h = v.Height()
	}

	borderStyle := text.Style{}.WithFg(text.BrightBlack())
	bgStyle := text.Style{}.WithBg(text.Color256(236))
	selStyle := text.Style{}.WithFg(text.Black()).WithBg(text.Cyan())
	itemStyle := text.Style{}.WithFg(text.White()).WithBg(text.Color256(236))
	shortcutStyle := text.Style{}.WithFg(text.BrightBlack()).WithBg(text.Color256(236))
	sepStyle := text.Style{}.WithFg(text.BrightBlack()).WithBg(text.Color256(236))

	// Fill background.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v.SetRune(x, y, ' ', bgStyle)
		}
	}

	// Top border.
	v.SetRune(0, 0, '┌', borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, 0, '─', borderStyle)
	}
	v.SetRune(w-1, 0, '┐', borderStyle)

	// Bottom border.
	v.SetRune(0, h-1, '└', borderStyle)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, h-1, '─', borderStyle)
	}
	v.SetRune(w-1, h-1, '┘', borderStyle)

	// Items.
	innerW := w - 2
	for i, it := range d.items {
		y := i + 1
		if y >= h-1 {
			break
		}
		v.SetRune(0, y, '│', borderStyle)
		v.SetRune(w-1, y, '│', borderStyle)

		if it.IsSep {
			v.SetRune(0, y, '├', borderStyle)
			for x := 1; x < w-1; x++ {
				v.SetRune(x, y, '─', sepStyle)
			}
			v.SetRune(w-1, y, '┤', borderStyle)
			continue
		}

		style := itemStyle
		scStyle := shortcutStyle
		if i == d.selected {
			style = selStyle
			scStyle = selStyle
		}

		// Fill row background.
		for x := 1; x < w-1; x++ {
			v.SetRune(x, y, ' ', style)
		}

		// Label.
		v.WriteString(2, y, it.Label, style)

		// Shortcut right-aligned.
		if it.Shortcut != "" {
			scLen := len([]rune(it.Shortcut))
			scX := 1 + innerW - scLen - 1
			if scX > 2+len([]rune(it.Label)) {
				v.WriteString(scX, y, it.Shortcut, scStyle)
			}
		}
	}
}

func (d *DropdownRenderable) HandleEvent(ev input.Event) bool {
	// This is handled by the MenuBar for keyboard, and the overlay system
	// calls back into MenuBar for mouse. This handler is for mouse clicks
	// dispatched by the overlay system.
	return false
}

// MenuBar is the horizontal menu bar at the top row.
type MenuBar struct {
	menus     []Menu
	openIdx   int // -1 = no menu open.
	dropdown  *DropdownRenderable
	onAction  func(menuTitle, itemLabel string)
	barBounds render.Rect // set by the layout system
}

func NewMenuBar(menus []Menu, onAction func(string, string)) *MenuBar {
	return &MenuBar{
		menus:    menus,
		openIdx:  -1,
		onAction: onAction,
	}
}

// menuX returns the x-offset where the given menu title starts in the bar.
func (mb *MenuBar) menuX(idx int) int {
	x := 1 // left padding
	for i := 0; i < idx; i++ {
		x += len([]rune(mb.menus[i].Title)) + 2 // title + spacing
	}
	return x
}

func (mb *MenuBar) openMenu(idx int) {
	if idx < 0 || idx >= len(mb.menus) {
		return
	}
	mb.openIdx = idx
	mb.dropdown = NewDropdownRenderable(mb.menus[idx].Items)
}

func (mb *MenuBar) closeMenu() {
	mb.openIdx = -1
	mb.dropdown = nil
}

func (mb *MenuBar) selectItem() {
	if mb.openIdx < 0 || mb.dropdown == nil {
		return
	}
	sel := mb.dropdown.selected
	items := mb.menus[mb.openIdx].Items
	if sel >= 0 && sel < len(items) && !items[sel].IsSep {
		menuTitle := mb.menus[mb.openIdx].Title
		itemLabel := items[sel].Label
		mb.closeMenu()
		if mb.onAction != nil {
			mb.onAction(menuTitle, itemLabel)
		}
	}
}

func (mb *MenuBar) moveSelection(delta int) {
	if mb.dropdown == nil {
		return
	}
	items := mb.dropdown.items
	n := len(items)
	if n == 0 {
		return
	}
	sel := mb.dropdown.selected
	for range n {
		sel += delta
		if sel < 0 {
			sel = n - 1
		} else if sel >= n {
			sel = 0
		}
		if !items[sel].IsSep {
			break
		}
	}
	mb.dropdown.selected = sel
}

func (mb *MenuBar) Render(v *render.View) {
	w := v.Width()
	barStyle := text.Style{}.WithFg(text.White()).WithBg(text.Color256(238))
	activeStyle := text.Style{}.WithFg(text.Black()).WithBg(text.Cyan())

	// Fill bar background.
	for x := 0; x < w; x++ {
		v.SetRune(x, 0, ' ', barStyle)
	}

	// Draw menu titles.
	x := 1
	for i, m := range mb.menus {
		style := barStyle
		if i == mb.openIdx {
			style = activeStyle
		}
		label := " " + m.Title + " "
		v.WriteString(x, 0, label, style)
		x += len([]rune(label))
	}
}

func (mb *MenuBar) HandleEvent(ev input.Event) bool {
	if ev.Type == input.EventKey {
		ke := ev.Key

		// Alt+letter opens menu.
		if ke.Mod == input.ModAlt && ke.Code == input.KeyRune {
			r := unicode.ToLower(ke.Rune)
			for i, m := range mb.menus {
				if len(m.Title) > 0 && unicode.ToLower(rune(m.Title[0])) == r {
					if mb.openIdx == i {
						mb.closeMenu()
					} else {
						mb.openMenu(i)
					}
					return true
				}
			}
		}

		// When menu is open, consume keyboard.
		if mb.openIdx >= 0 {
			switch ke.Code {
			case input.KeyEscape:
				mb.closeMenu()
				return true
			case input.KeyUp:
				mb.moveSelection(-1)
				return true
			case input.KeyDown:
				mb.moveSelection(1)
				return true
			case input.KeyLeft:
				idx := mb.openIdx - 1
				if idx < 0 {
					idx = len(mb.menus) - 1
				}
				mb.openMenu(idx)
				return true
			case input.KeyRight:
				idx := (mb.openIdx + 1) % len(mb.menus)
				mb.openMenu(idx)
				return true
			case input.KeyEnter:
				mb.selectItem()
				return true
			}
			// Consume all other keys while menu is open.
			return true
		}
	}

	if ev.Type == input.EventMouse {
		m := ev.Mouse
		if m.Button == input.MouseLeft {
			// Click on the bar row?
			if m.Y == 0 {
				// Find which menu title was clicked.
				x := 1
				for i, menu := range mb.menus {
					label := " " + menu.Title + " "
					labelW := len([]rune(label))
					if m.X >= x && m.X < x+labelW {
						if mb.openIdx == i {
							mb.closeMenu()
						} else {
							mb.openMenu(i)
						}
						return true
					}
					x += labelW
				}
				// Clicked on bar but not on a title: close.
				mb.closeMenu()
				return true
			}
		}
	}

	return false
}

func (mb *MenuBar) Overlays() []render.OverlayRequest {
	if mb.openIdx < 0 || mb.dropdown == nil {
		return nil
	}
	x := mb.menuX(mb.openIdx)
	return []render.OverlayRequest{{
		Renderable: &dropdownInteractive{bar: mb, dd: mb.dropdown},
		Anchor: render.Rect{
			X:      x,
			Y:      1, // Below the bar row.
			Width:  mb.dropdown.width,
			Height: mb.dropdown.height,
		},
		ZOrder: 100,
	}}
}

// dropdownInteractive wraps the dropdown renderable so overlay mouse dispatch
// can route clicks back into the menu bar's logic.
type dropdownInteractive struct {
	bar *MenuBar
	dd  *DropdownRenderable
}

func (di *dropdownInteractive) Render(v *render.View) {
	di.dd.Render(v)
}

func (di *dropdownInteractive) HandleEvent(ev input.Event) bool {
	if ev.Type == input.EventMouse && ev.Mouse.Button == input.MouseLeft {
		mx, my := ev.Mouse.X, ev.Mouse.Y
		// Items start at y=1 inside the dropdown.
		itemIdx := my - 1
		if itemIdx >= 0 && itemIdx < len(di.dd.items) && mx >= 1 && mx < di.dd.width-1 {
			if !di.dd.items[itemIdx].IsSep {
				di.dd.selected = itemIdx
				di.bar.selectItem()
				return true
			}
		}
		// Click inside dropdown but not on an item: consume.
		return true
	}
	return false
}

// ContentArea displays a scrollable log of menu actions.
type ContentArea struct {
	lines   []text.StyledLine
	maxLog  int
	focused bool
	scrollY int
}

func NewContentArea() *ContentArea {
	c := &ContentArea{maxLog: 200}
	c.addLine(text.Parse("[bold cyan]Menu Bar Demo[/]"))
	c.addLine(text.Parse("[dim]Click menu titles or press Alt+first letter to open menus.[/]"))
	c.addLine(text.Parse("[dim]Use arrow keys to navigate, Enter to select, Escape to close.[/]"))
	c.addLine(text.Parse(""))
	return c
}

func (c *ContentArea) addLine(line text.StyledLine) {
	c.lines = append(c.lines, line)
	if len(c.lines) > c.maxLog {
		c.lines = c.lines[len(c.lines)-c.maxLog:]
	}
	// Auto-scroll to bottom.
	c.scrollY = len(c.lines)
}

func (c *ContentArea) AddAction(menuTitle, itemLabel string) {
	msg := fmt.Sprintf("[bold green]>[/] [bold]%s[/] [dim]>[/] [yellow]%s[/] triggered", menuTitle, itemLabel)
	c.addLine(text.Parse(msg))
}

func (c *ContentArea) Render(v *render.View) {
	w, h := v.Width(), v.Height()
	bgStyle := text.Style{}.WithBg(text.Color256(234))

	// Fill background.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v.SetRune(x, y, ' ', bgStyle)
		}
	}

	// Determine visible lines.
	visibleH := h
	start := 0
	if len(c.lines) > visibleH {
		start = len(c.lines) - visibleH
	}
	if c.scrollY < len(c.lines) && c.scrollY-visibleH > start {
		start = c.scrollY - visibleH
		if start < 0 {
			start = 0
		}
	}

	for i, line := range c.lines[start:] {
		if i >= visibleH {
			break
		}
		col := 1
		for _, span := range line {
			style := span.Style
			if style.Bg.Equal(text.Default()) {
				style.Bg = text.Color256(234)
			}
			col += v.WriteString(col, i, span.Text, style)
		}
	}
}

func (c *ContentArea) HandleEvent(ev input.Event) bool {
	if ev.Type == input.EventMouse {
		if ev.Mouse.Button == input.MouseLeft {
			return true
		}
	}
	return false
}

func (c *ContentArea) Focus()          { c.focused = true }
func (c *ContentArea) Blur()           { c.focused = false }
func (c *ContentArea) IsFocused() bool { return c.focused }

// StatusBar displays a status line at the bottom.
type StatusBar struct {
	text string
}

func NewStatusBar() *StatusBar {
	return &StatusBar{text: " Alt+F/E/V/H: open menus | Arrow keys: navigate | Enter: select | Esc: close | Ctrl-C: quit"}
}

func (sb *StatusBar) Render(v *render.View) {
	w := v.Width()
	style := text.Style{}.WithFg(text.Black()).WithBg(text.White())

	for x := 0; x < w; x++ {
		v.SetRune(x, 0, ' ', style)
	}

	display := sb.text
	if len([]rune(display)) > w {
		display = string([]rune(display)[:w])
	}
	v.WriteString(0, 0, display, style)
}

// AppRoot wraps the split layout and provides overlay pass-through from MenuBar.
type AppRoot struct {
	split   *layout.Split
	menuBar *MenuBar
	bounds  render.Rect
}

func (ar *AppRoot) Layout(bounds render.Rect) {
	ar.bounds = bounds
	ar.split.Layout(bounds)
}

func (ar *AppRoot) Render(v *render.View) {
	ar.split.Render(v)
}

func (ar *AppRoot) HandleEvent(ev input.Event) bool {
	return ar.split.HandleEvent(ev)
}

func (ar *AppRoot) Children() []render.Renderable {
	return ar.split.Children()
}

func (ar *AppRoot) ChildBounds() []render.Rect {
	return ar.split.ChildBounds()
}

func (ar *AppRoot) Overlays() []render.OverlayRequest {
	return nil // Overlays come from MenuBar via the container tree.
}

func main() {
	t, err := term.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()

	content := NewContentArea()
	statusBar := NewStatusBar()

	menus := []Menu{
		{
			Title: "File",
			Items: []MenuItem{
				{Label: "New"},
				{Label: "Open"},
				{Label: "Save", Shortcut: "Ctrl-S"},
				{Label: "Save As"},
				{IsSep: true},
				{Label: "Quit", Shortcut: "Ctrl-C"},
			},
		},
		{
			Title: "Edit",
			Items: []MenuItem{
				{Label: "Undo", Shortcut: "Ctrl-Z"},
				{Label: "Redo"},
				{IsSep: true},
				{Label: "Cut"},
				{Label: "Copy"},
				{Label: "Paste"},
			},
		},
		{
			Title: "View",
			Items: []MenuItem{
				{Label: "Zoom In"},
				{Label: "Zoom Out"},
				{IsSep: true},
				{Label: "Toggle Sidebar"},
			},
		},
		{
			Title: "Help",
			Items: []MenuItem{
				{Label: "About"},
				{Label: "Shortcuts"},
			},
		},
	}

	menuBar := NewMenuBar(menus, func(menuTitle, itemLabel string) {
		content.AddAction(menuTitle, itemLabel)
		if menuTitle == "File" && itemLabel == "Quit" {
			// Will be handled by the app quit binding.
		}
	})

	split := layout.NewVSplit(
		layout.SplitChild{Renderable: menuBar, Size: 1},
		layout.SplitChild{Renderable: content, Size: 1.0},
		layout.SplitChild{Renderable: statusBar, Size: 1},
	)

	root := &AppRoot{
		split:   split,
		menuBar: menuBar,
	}

	app := layout.NewApp(t, root)
	app.SetFocus(content)

	// Ctrl-C quits.
	app.BindRune('c', input.ModCtrl, func() { app.Quit() })

	if err := app.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
