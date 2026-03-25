package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/iceisfun/gorepl/pkg/hexedit"
	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/layout"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/term"
	"github.com/iceisfun/gorepl/pkg/text"
)

// HexApp ties together the hex view, info panel, status bar, and modal dialogs.
type HexApp struct {
	hexView  *hexedit.HexView
	info     *hexedit.InfoPanel
	search   *hexedit.SearchDialog
	gotoDlg  *hexedit.GotoDialog
	statusFn func() string // Provides status text from the app.
	app      *layout.App

	lastSearchPattern []byte
	lastSearchMode    hexedit.SearchMode
	lastSearchAlign   hexedit.Alignment
}

// StatusBar renders a single-line status bar at the bottom.
type StatusBar struct {
	getText func() string
	style   text.Style
}

func (s *StatusBar) Render(v *render.View) {
	w := v.Width()
	msg := s.getText()
	for x := range w {
		v.SetRune(x, 0, ' ', s.style)
	}
	v.WriteString(0, 0, msg, s.style)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: hexedit <file>\n")
		os.Exit(1)
	}

	buf, err := hexedit.OpenFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	t, err := term.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()

	hv := hexedit.NewHexView(buf)
	info := hexedit.NewInfoPanel(hv)
	searchDlg := hexedit.NewSearchDialog()
	gotoDlg := hexedit.NewGotoDialog()

	ha := &HexApp{
		hexView: hv,
		info:    info,
		search:  searchDlg,
		gotoDlg: gotoDlg,
	}

	searchDlg.OnAction = func(action string) {
		if action == "search" {
			ha.doSearch()
		}
	}
	gotoDlg.OnAction = func(action string) {
		if action == "goto" {
			ha.doGoto()
		}
	}

	statusBar := &StatusBar{
		style: text.Style{}.WithFg(text.Black()).WithBg(text.Color256(240)),
		getText: func() string {
			cursor := hv.Cursor()
			row := cursor/16 + 1
			totalRows := hv.TotalRows()
			pct := 0
			if totalRows > 0 {
				pct = row * 100 / totalRows
			}
			return fmt.Sprintf(" %s | %s | Offset: 0x%08X (%d) | Row %d/%d (%d%%) | Ctrl-F:Search Ctrl-G:Goto Ctrl-C:Quit",
				buf.Path(), hexedit.FormatSize(buf.Len()), cursor, cursor, row, totalRows, pct)
		},
	}
	ha.statusFn = statusBar.getText

	// Layout: hex view (left) | info panel (right), status bar at bottom.
	hexAndInfo := layout.NewHSplit(
		layout.SplitChild{Renderable: hv, Size: 0.6},
		layout.SplitChild{Renderable: info, Size: 0.4},
	)

	root := layout.NewVSplit(
		layout.SplitChild{Renderable: hexAndInfo, Size: 1.0},
		layout.SplitChild{Renderable: statusBar, Size: 2}, // Fixed 1 row + border.
	)

	app := layout.NewApp(t, root)
	ha.app = app
	app.SetFocus(hv)

	// --- Key bindings ---

	// Quit.
	app.BindRune('c', input.ModCtrl, func() { app.Quit() })

	// Search (Ctrl-F or /).
	openSearch := func() {
		searchDlg.Show()
		app.RequestRender()
	}
	app.BindRune('f', input.ModCtrl, openSearch)
	app.BindRune('/', 0, openSearch)

	// Find next (Ctrl-N or n).
	app.BindRune('n', input.ModCtrl, func() { ha.searchNext() })

	// Goto (Ctrl-G).
	app.BindRune('g', input.ModCtrl, func() {
		gotoDlg.Show()
		app.RequestRender()
	})

	// Jump (Ctrl-J) — same as goto.
	app.BindRune('j', input.ModCtrl, func() {
		gotoDlg.Show()
		app.RequestRender()
	})

	// Override the app's dispatch to handle modal dialogs.
	origRun := app.Run
	_ = origRun // We'll wrap the event handling instead.

	// We need to intercept events when dialogs are visible.
	// Use a wrapper renderable that intercepts events.
	wrapper := &DialogInterceptor{
		child:     root,
		hexApp:    ha,
		searchDlg: searchDlg,
		gotoDlg:   gotoDlg,
	}

	// Rebuild with the wrapper as root.
	app2 := layout.NewApp(t, wrapper)
	ha.app = app2
	app2.SetFocus(hv)

	app2.BindRune('c', input.ModCtrl, func() { app2.Quit() })
	app2.BindRune('f', input.ModCtrl, func() {
		searchDlg.Show()
		app2.RequestRender()
	})
	app2.BindRune('/', 0, func() {
		if !searchDlg.IsVisible() && !gotoDlg.IsVisible() {
			searchDlg.Show()
			app2.RequestRender()
		}
	})
	app2.BindRune('n', input.ModCtrl, func() { ha.searchNext() })
	app2.BindRune('g', input.ModCtrl, func() {
		gotoDlg.Show()
		app2.RequestRender()
	})
	app2.BindRune('j', input.ModCtrl, func() {
		gotoDlg.Show()
		app2.RequestRender()
	})

	if err := app2.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (ha *HexApp) doSearch() {
	pattern, err := hexedit.ParseSearchPattern(ha.search.Input(), ha.search.Mode())
	if err != nil {
		ha.search.SetMessage(err.Error())
		return
	}

	ha.lastSearchPattern = pattern
	ha.lastSearchMode = ha.search.Mode()
	ha.lastSearchAlign = ha.search.Align()

	align := int(ha.lastSearchAlign)
	result := ha.hexView.Buffer().SearchAligned(pattern, ha.hexView.Cursor(), align)
	if result < 0 {
		// Wrap around from beginning.
		result = ha.hexView.Buffer().SearchAligned(pattern, 0, align)
	}
	if result < 0 {
		ha.search.SetMessage("Not found")
		return
	}

	ha.hexView.SetCursor(result)
	ha.hexView.SetHighlight(result, len(pattern))
	ha.search.Hide()
}

func (ha *HexApp) searchNext() {
	if len(ha.lastSearchPattern) == 0 {
		return
	}
	align := int(ha.lastSearchAlign)
	result := ha.hexView.Buffer().SearchAligned(ha.lastSearchPattern, ha.hexView.Cursor()+1, align)
	if result < 0 {
		result = ha.hexView.Buffer().SearchAligned(ha.lastSearchPattern, 0, align)
	}
	if result >= 0 {
		ha.hexView.SetCursor(result)
		ha.hexView.SetHighlight(result, len(ha.lastSearchPattern))
	}
}

func (ha *HexApp) doGoto() {
	input := ha.gotoDlg.Input()
	fileLen := ha.hexView.Buffer().Len()

	offset, err := ha.gotoDlg.ParseOffset()
	if err != nil {
		ha.gotoDlg.SetMessage("Invalid offset")
		return
	}

	// Relative offset: if starts with + or -, apply relative to cursor.
	trimmed := strings.TrimSpace(input)
	if len(trimmed) > 0 && (trimmed[0] == '+' || trimmed[0] == '-') {
		offset = int64(ha.hexView.Cursor()) + offset
	}

	// Clamp to file bounds instead of erroring.
	if offset < 0 {
		offset = 0
	}
	if int(offset) >= fileLen {
		offset = int64(fileLen - 1)
	}

	ha.hexView.GotoOffset(int(offset))
	ha.hexView.ClearHighlight()
	ha.gotoDlg.Hide()
}

// DialogInterceptor wraps the main layout and intercepts events
// when a modal dialog is visible.
type DialogInterceptor struct {
	child     render.Renderable
	hexApp    *HexApp
	searchDlg *hexedit.SearchDialog
	gotoDlg   *hexedit.GotoDialog
}

func (d *DialogInterceptor) Render(v *render.View) {
	d.child.Render(v)
}

func (d *DialogInterceptor) HandleEvent(ev input.Event) bool {
	// Modal dialogs consume ALL keyboard/paste events.
	// Mouse events are dispatched by the App's overlay hit-testing, so
	// they arrive here only if they missed the overlay — consume them
	// to prevent leaking to the view behind.
	if d.searchDlg.IsVisible() {
		if ev.Type == input.EventMouse {
			return true // Consumed — overlay dispatch handles clicks inside the dialog.
		}
		d.searchDlg.HandleEvent(ev)
		return true
	}

	if d.gotoDlg.IsVisible() {
		if ev.Type == input.EventMouse {
			return true
		}
		d.gotoDlg.HandleEvent(ev)
		return true
	}

	// Pass to child layout.
	if h, ok := d.child.(render.Interactive); ok {
		return h.HandleEvent(ev)
	}
	return false
}

// Forward Container interface for layout to work.
func (d *DialogInterceptor) Layout(bounds render.Rect) {
	if c, ok := d.child.(interface{ Layout(render.Rect) }); ok {
		c.Layout(bounds)
	}
}

func (d *DialogInterceptor) Children() []render.Renderable {
	if c, ok := d.child.(render.Container); ok {
		return c.Children()
	}
	return []render.Renderable{d.child}
}

func (d *DialogInterceptor) ChildBounds() []render.Rect {
	if c, ok := d.child.(render.Container); ok {
		return c.ChildBounds()
	}
	return nil
}

// Overlayable — render dialogs as overlays.
func (d *DialogInterceptor) Overlays() []render.OverlayRequest {
	var overlays []render.OverlayRequest

	if d.searchDlg.IsVisible() {
		sw, sh := d.searchDlg.Size()
		overlays = append(overlays, render.OverlayRequest{
			Renderable: d.searchDlg,
			Anchor:     render.Rect{X: 4, Y: 2, Width: sw, Height: sh},
			ZOrder:     200,
		})
	}

	if d.gotoDlg.IsVisible() {
		gw, gh := d.gotoDlg.Size()
		overlays = append(overlays, render.OverlayRequest{
			Renderable: d.gotoDlg,
			Anchor:     render.Rect{X: 4, Y: 2, Width: gw, Height: gh},
			ZOrder:     200,
		})
	}

	return overlays
}
