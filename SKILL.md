---
name: gotui
description: Build terminal user interfaces in Go using github.com/iceisfun/gotui. Covers the render pipeline, layout system, event dispatch, widget library, overlay compositing, and the embeddable REPL component.
license: MIT
compatibility: claude-code, opencode
metadata:
  language: go
  domain: terminal-ui
---

# GoREPL TUI Framework -- Skill Reference

Use this when helping someone who imported `github.com/iceisfun/gotui` and wants to build terminal applications in Go.

## SKILLS

Copy-paste block for an AI assistant:

```text
SKILLS:
- GoREPL is a composable TUI framework for Go with a cell-based render pipeline.
- Most apps only need five steps: term.Open -> build root Renderable -> layout.NewApp -> app.Run -> term.Close.
- Everything draws by implementing Render(v *render.View). Capabilities are added progressively via type assertions (Interactive, Focusable, Scrollable, Cursorable, Overlayable, Updatable, Container).
- Key architectural rules:
  - All coordinates passed to a Renderable are local (view-relative).
  - Mouse events carry local X,Y (translated by the layout system) and absolute ScreenX,ScreenY.
  - Overlays render into a separate transparent buffer and are composited on top.
  - Event dispatch order: global bindings -> overlay hit-test -> focused renderable -> bubble up through containers.
  - Tab/Backtab auto-cycle focus by default.
- Cell has Rune, Width, Style, and Transparent flag. Transparent cells fall through during compositing.
- View provides SetCell, SetRune, WriteString, Fill, Clear, ClearTransparent, Sub.
- Layout containers: Split (HSplit/VSplit with ratio or fixed sizes), Absolute, Viewport, Modal.
- SplitChild.Size: <=1.0 is a ratio, >1.0 is fixed cell count.
- Widgets: Label, Button, Selector, List, Tree, Menu, Panel, TextView.
- Panel draws a border (Single/Double/Rounded) and delegates to a child.
- text.Parse("[bold red]hello[/]") returns a StyledLine with markup support.
- Color constructors: text.Red(), text.RGB(r,g,b), text.Color256(n). Style chains: style.Bold().WithFg(text.Cyan()).
- The repl package provides a full REPL component (Editor, History, CompletionPopup, Output) that implements Focusable, Cursorable, Scrollable, and Overlayable. It handles paste events automatically.
- REPL needs an Executor interface: Execute(source) (string, error) and IsComplete(source) bool. Optional: Completer, Highlighter.
- Paste events: check ev.Type == input.EventPaste before EventKey. The pasted text is in ev.Paste. Call InsertString(ev.Paste) to insert it.
```

## Smallest Useful Example

```go
package main

import (
    "context"
    "log"

    "github.com/iceisfun/gotui/pkg/input"
    "github.com/iceisfun/gotui/pkg/layout"
    "github.com/iceisfun/gotui/pkg/render"
    "github.com/iceisfun/gotui/pkg/term"
    "github.com/iceisfun/gotui/pkg/text"
)

type Hello struct{}

func (h *Hello) Render(v *render.View) {
    v.WriteString(1, 1, "Hello, TUI!", text.Style{}.Bold().WithFg(text.Cyan()))
    v.WriteString(1, 2, "Press Ctrl+C to quit.", text.Style{}.Dim())
}

func main() {
    t, err := term.Open()
    if err != nil {
        log.Fatal(err)
    }
    defer t.Close()

    app := layout.NewApp(t, &Hello{})
    app.BindRune('c', input.ModCtrl, func() { app.Quit() })

    if err := app.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## Architecture

### Package Dependency Diagram

```
text  (Color, Style, StyledSpan, StyledLine, Parse, Wrap)
  |
term  (Terminal, Writer, Reader -- raw mode, ANSI output, background input)
  |
input (Event, KeyEvent, MouseEvent, Decoder -- byte stream to events)
  |
render (Cell, Buffer, View, Rect, Composite, Flush, Surface, interfaces)
  |
layout (App, Split, Absolute, Viewport, Modal, overlay collector)
  |
widget (Label, Button, Selector, List, Tree, Menu, Panel, TextView)
  |
repl   (Executor, Completer, Highlighter, Editor, History, CompletionPopup, REPL)
hexedit (Buffer, HexView, InfoPanel, SearchDialog, GotoDialog)
```

### Event Flow

```
Terminal stdin -> Reader (background goroutine) -> raw []byte channel
  -> App.Run select loop -> Decoder.Feed(raw) -> []input.Event
  -> App.dispatch:
     1. Global KeyBindings (checked first, short-circuit if matched)
     2. Mouse events: overlay hit-test (topmost z-order first) -> container tree hit-test (deepest child first)
     3. Key/Paste events: focused Focusable -> bubble to root Interactive
```

### Render Pipeline

```
App.render():
  1. baseBuf.Clear()
  2. root.Render(baseView)                -- normal tree renders into base buffer
  3. CollectOverlays(root) + SortOverlays -- gather overlay requests, sort by z-order
  4. overlayBuf.ClearTransparent()
  5. Each overlay renders into overlayBuf  -- transparent cells fall through
  6. CompositeTwo(back, baseBuf, overlayBuf)
  7. Flush(writer, front, back)            -- diff prev vs curr, emit only changed cells
  8. Position hardware cursor via Cursorable on focused renderable
```

## Core Concepts

### Renderable Interface Hierarchy

```go
Renderable          { Render(v *View) }
  Interactive       { HandleEvent(ev Event) bool }       // +input handling
    Focusable       { Focus(); Blur(); IsFocused() bool } // +keyboard focus
  Scrollable        { ContentSize() (w,h); ScrollOffset() (x,y); SetScrollOffset(x,y) }
  Cursorable        { CursorPosition() (x, y int, visible bool) }
  Overlayable       { Overlays() []OverlayRequest }
  Updatable         { Update(dt time.Duration) }
  Measurable        { MinSize(); MaxSize(); PreferredSize() }
  Container         { Children() []Renderable; Layout(Rect); ChildBounds() []Rect }
  MouseTarget       { HitTest(x, y int) bool }
```

Capabilities are discovered via type assertion. A widget implements only what it needs. The App and layout system check `if f, ok := r.(render.Focusable); ok { ... }` at runtime.

### Cell and Transparency

```go
type Cell struct {
    Rune        rune
    Width       uint8       // 1 for ASCII, 2 for CJK, 0 for continuation cell
    Style       text.Style
    Transparent bool        // true = falls through to the layer below
}
var BlankCell       = Cell{Rune: ' ', Width: 1}
var TransparentCell = Cell{Transparent: true}
```

Overlays use `ClearTransparent()` so only the cells they write are visible; everything else shows the base layer underneath.

### Layout System

- **App** -- owns the terminal, event loop, focus, global bindings, and compositing. Calls `Layout(fullScreenRect)` on root, then `Render`.
- **Split** -- divides space along an axis. `NewHSplit(children...)` / `NewVSplit(children...)`. Each `SplitChild{Renderable, Size}` where Size <=1.0 is a ratio and >1.0 is fixed cells.
- **Absolute** -- positions children at explicit `Rect` coordinates.
- **Viewport** -- wraps a child with scroll offset. Handles mouse wheel. Forwards `SetScrollOffset` to Scrollable children.
- **Modal** -- when `Show()`'d, produces an `OverlayRequest` centered on screen (z-order 100). Consumes all events while visible.

### Overlay System

Any renderable implementing `Overlayable` can return `[]OverlayRequest`. Each request has a `Renderable`, `Anchor` rect (in local coords), and `ZOrder`. The layout system:
1. Walks the tree with `CollectOverlays`, translating anchors to absolute screen coordinates.
2. Sorts by z-order ascending.
3. Renders all overlays into a transparent buffer.
4. Composites: topmost non-transparent cell wins.
5. Mouse dispatch checks overlays first (highest z-order first).

## Widget Library

| Widget | Constructor | Interfaces | Usage |
|--------|------------|------------|-------|
| **Label** | `NewLabel(s)` / `NewStyledLabel(line)` | Renderable | `l := widget.NewLabel("Status: OK")` / `l.Align = widget.AlignCenter` |
| **Button** | `NewButton(label, onClick)` | Focusable | `b := widget.NewButton("Save", func() { save() })` |
| **Selector** | `NewSelector(items)` | Focusable | `s := widget.NewSelector([]string{"A","B","C"})` / `s.OnChange = func(i int) {}` |
| **List** | `NewList(items)` | Focusable, Scrollable | `l := widget.NewList(names)` / `l.OnSelect = func(i int) {}` |
| **Tree** | `NewTree(roots)` | Focusable, Scrollable | `t := widget.NewTree([]*widget.TreeNode{{Label: "Root", Children: ...}})` |
| **Menu** | `NewMenu(items)` | Interactive | `m := widget.NewMenu([]widget.MenuItem{{Label: "Open", Action: open}})` / `m.Show()` |
| **Panel** | `NewPanel(title, child)` | Container | `p := widget.NewPanel("Info", myWidget)` / `p.Border = widget.BorderRounded` |
| **TextView** | `NewTextView()` | Interactive, Scrollable | `tv := widget.NewTextView()` / `tv.AppendPlain("line")` / `tv.AutoScroll = true` |

## Common Patterns

### Creating a Custom Renderable

```go
type StatusBar struct {
    Message string
    style   text.Style
}

func (s *StatusBar) Render(v *render.View) {
    v.Fill(render.Rect{X: 0, Y: 0, Width: v.Width(), Height: 1},
        render.Cell{Rune: ' ', Width: 1, Style: s.style})
    v.WriteString(1, 0, s.Message, s.style)
}
```

### Handling Mouse Clicks with Hit Regions

Record regions during Render, check during HandleEvent:

```go
type ClickableList struct {
    items   []string
    regions []render.Rect // populated each Render
}

func (c *ClickableList) Render(v *render.View) {
    c.regions = c.regions[:0]
    for i, item := range c.items {
        v.WriteString(0, i, item, text.Style{})
        c.regions = append(c.regions, render.Rect{X: 0, Y: i, Width: v.Width(), Height: 1})
    }
}

func (c *ClickableList) HandleEvent(ev input.Event) bool {
    if ev.Type == input.EventMouse && ev.Mouse.Button == input.MouseLeft {
        for i, r := range c.regions {
            if r.Contains(ev.Mouse.X, ev.Mouse.Y) {
                fmt.Println("clicked", c.items[i])
                return true
            }
        }
    }
    return false
}
```

### Handling Paste Events

The terminal enables bracketed paste mode automatically. The decoder parses
`ESC[200~...ESC[201~` sequences into `EventPaste` events with the pasted text
in `ev.Paste`. Handle them before checking for key events:

```go
func (e *MyEditor) HandleEvent(ev input.Event) bool {
    if ev.Type == input.EventPaste {
        e.InsertString(ev.Paste)
        return true
    }
    if ev.Type != input.EventKey {
        return false
    }
    // ... handle key events
}
```

The built-in REPL component handles paste events automatically via the editor's
`InsertString` method.

### Creating a Modal Dialog

```go
content := widget.NewLabel("Are you sure?")
dialog := layout.NewModal(content, 40, 5)
dialog.SetScreenSize(cols, rows) // or let Layout set it

// In your app setup:
app.BindRune('d', input.ModCtrl, func() { dialog.Show() })

// Include the modal in your renderable tree so overlays are collected:
root := layout.NewVSplit(
    layout.SplitChild{Renderable: mainContent, Size: 1.0},
    layout.SplitChild{Renderable: dialog, Size: 0}, // zero-height; renders via overlay
)
```

### Using the Markup System

Syntax: `[attrs]text[/]` where attrs are space-separated tokens.

```go
line := text.Parse("[bold red]Error:[/] [dim]file not found[/]")
line2 := text.Parse("[italic #ff8800]warning[/] [on #003366]highlighted bg[/]")
// Supported attrs: bold, dim, italic, underline, blink, reverse, strikethrough
// Colors: red, green, blue, yellow, cyan, magenta, white, black, bright_red, etc.
// "on <color>" sets background. #RRGGBB for RGB. Tags nest with [/] popping one level.
```

### Split Layout with Status Bar

```go
root := layout.NewVSplit(
    layout.SplitChild{Renderable: mainContent, Size: 1.0},  // fills remaining space
    layout.SplitChild{Renderable: statusBar, Size: 3},       // fixed 3 rows at bottom
)
app := layout.NewApp(t, root)
```

### Using the REPL Component

```go
type MyExecutor struct{}
func (e *MyExecutor) Execute(src string) (string, error) { return eval(src), nil }
func (e *MyExecutor) IsComplete(src string) bool         { return !strings.HasSuffix(src, "\\") }

r := repl.New(&MyExecutor{},
    repl.WithPrompt(">> ", ".. "),
    repl.WithHighlighter(myHighlighter),
    repl.WithCompleter(myCompleter),
    repl.WithHistorySize(500),
)
// r implements Renderable, Focusable, Cursorable, Scrollable, Overlayable
app := layout.NewApp(t, r)
app.SetFocus(r)
```

## Guidance for AI Assistants

Key rules when helping users build with this framework:

1. **Use the five-step flow**: `term.Open` -> build renderable tree -> `layout.NewApp` -> bind keys / set focus -> `app.Run`. Always defer `term.Close()`.

2. **Implement interfaces progressively**. Start with `Render(v *View)`. Add `HandleEvent` only if the widget needs input. Add `Focus/Blur/IsFocused` only if it should receive keyboard focus.

3. **Coordinates are always local**. A Renderable receives a View whose (0,0) is its own top-left. Mouse event X,Y are translated to local coords by the layout system. Use ScreenX,ScreenY only for absolute positioning.

4. **Use transparency for overlays**. Call `v.ClearTransparent()` at the start of overlay Render methods. Only write the cells that should be visible; everything else shows through.

5. **SplitChild.Size semantics**: values <= 1.0 are ratios (proportional), values > 1.0 are fixed cell counts. Use fixed sizes for status bars and toolbars, ratios for content areas.

6. **Event consumption**: `HandleEvent` returns `true` to consume (stop propagation) or `false` to bubble. Global bindings on App are checked before any renderable.

7. **Focus management**: Call `app.SetFocus(f)` to focus a Focusable. Use `app.CollectFocusables()` to auto-populate the tab order from the tree. Tab/Backtab cycle focus automatically.

8. **Hit regions pattern**: Record clickable regions during Render, check them in HandleEvent. Mouse coords are local to the renderable.

9. **Modal dialogs**: Use `layout.Modal` for centered overlays. Include the Modal in the tree (even at zero size) so the overlay collector finds it. It consumes all events when visible.

10. **Prefer text.Parse for styled text**. The markup syntax `[bold red on blue]text[/]` is concise and nestable. Use `text.Style{}` chains for programmatic styling.
