# GoREPL -- Terminal Application Framework for Go

GoREPL is a pure-Go framework for building rich terminal user interfaces. It
provides a cell-based rendering engine with double-buffered differential updates,
a compositing overlay system with transparency, flexible layout containers,
input decoding for keyboard, mouse, and paste events, and a library of
ready-to-use widgets.

## Architecture

```
 cmd/demo  cmd/hexedit  cmd/dungeon  cmd/menu  cmd/fsview
     |          |            |           |          |
     +----------+-----+------+-----------+----------+
                      |
               pkg/layout          App, Split, Absolute, Viewport, Modal
                   |               Overlay collection, focus management
               pkg/widget          Label, Button, List, Tree, Menu, Panel, ...
                   |
               pkg/render          Buffer, Cell, View, Composite, Flush, Surface
                   |               Renderable, Interactive, Focusable, Container
               pkg/input           Decoder, Event, KeyCode, MouseEvent
                   |
               pkg/term            Terminal, Writer, Reader
                   |
               pkg/text            Color, Style, StyledLine, Parse, Wrap, Width
```

## Package Overview

| Package | Description | Key Types |
|---------|-------------|-----------|
| `pkg/text` | Terminal text styling, color management, markup parsing, width measurement, and word wrapping | `Color`, `Style`, `StyledLine`, `Parse`, `Wrap` |
| `pkg/term` | Raw terminal I/O with alternate screen, mouse tracking, and bracketed paste | `Terminal`, `Writer`, `Reader` |
| `pkg/input` | Decodes raw terminal bytes into structured keyboard, mouse, resize, and paste events | `Decoder`, `Event`, `KeyEvent`, `MouseEvent` |
| `pkg/render` | Cell-based framebuffer, clipped views, multi-layer compositing, and differential flush | `Buffer`, `Cell`, `View`, `Composite`, `Flush`, `Surface` |
| `pkg/layout` | Application shell with event loop, layout containers, focus cycling, and overlay management | `App`, `Split`, `Absolute`, `Viewport`, `Modal` |
| `pkg/widget` | Ready-to-use UI components | `Label`, `Button`, `Selector`, `List`, `Tree`, `Menu`, `Panel`, `TextView` |
| `pkg/repl` | Interactive read-eval-print loop with editor, history, completion, and syntax highlighting | `REPL`, `Editor`, `History`, `CompletionPopup` |
| `pkg/hexedit` | Hex editor components with data inspector, search, and goto dialogs | `HexView`, `Buffer`, `InfoPanel`, `SearchDialog`, `GotoDialog` |

## Quick Start

A minimal application that creates a terminal with two panels side by side:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/iceisfun/gotui/pkg/input"
    "github.com/iceisfun/gotui/pkg/layout"
    "github.com/iceisfun/gotui/pkg/term"
    "github.com/iceisfun/gotui/pkg/widget"
)

func main() {
    t, err := term.Open()
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    defer t.Close()

    left := widget.NewPanel("Left", widget.NewLabel("Hello"))
    right := widget.NewPanel("Right", widget.NewLabel("World"))

    root := layout.NewHSplit(
        layout.SplitChild{Renderable: left, Size: 0.5},
        layout.SplitChild{Renderable: right, Size: 0.5},
    )

    app := layout.NewApp(t, root)

    // Ctrl-C quits.
    app.BindRune('c', input.ModCtrl, func() { app.Quit() })

    if err := app.Run(context.Background()); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

## Features

- **Rendering engine**: Cell-based double-buffered framebuffer with differential
  flush that only writes changed cells to the terminal via ANSI escape sequences.
- **Compositing with transparency**: Multi-layer buffer compositing where
  transparent cells fall through to lower layers, enabling overlay effects.
- **Overlay system**: Renderables can produce floating overlay requests (popups,
  tooltips, dropdown menus) that are collected from the tree, sorted by z-order,
  and composited above base content.
- **Layout system**: Split (horizontal/vertical with proportional and fixed
  sizing), Absolute positioning, scrollable Viewport, and centered Modal
  containers.
- **Input handling**: Full keyboard decoding (UTF-8, escape sequences, CSI
  parameters, SS3, function keys, modifier detection), SGR mouse tracking
  (press, release, drag, motion, wheel), bracketed paste, and terminal resize
  via SIGWINCH.
- **Mouse hit-testing**: Automatic walk of the container tree to dispatch mouse
  events to the deepest matching renderable with correctly translated local
  coordinates. Overlays are checked first.
- **Focus management**: Tab/Backtab cycling through focusable components,
  mouse-click-to-focus, and global key binding registration.
- **Widget library**: Label, Button, Selector, List, Tree, Menu, Panel, and
  TextView with keyboard and mouse interaction.
- **Infinite scrollable surface**: Virtual canvas with camera positioning for
  applications like map viewers and game worlds.
- **Bracketed paste**: Paste events are decoded from `ESC[200~...ESC[201~`
  into `EventPaste` with the pasted text in `ev.Paste`. The REPL component
  handles paste automatically via the editor's `InsertString` method.
- **CJK-aware text handling**: Rune width measurement, string truncation, and
  word wrapping that correctly handles double-width characters.
- **Markup parser**: Bracket-tag syntax (`[bold red on blue]text[/]`) for
  inline styled text with nesting support.

## Example Applications

### cmd/demo -- Click Hit-Test Demo

Two side-by-side panels with clickable colored buttons. Demonstrates mouse
hit-testing through the layout tree, overlay compositing (toggle with Ctrl-O),
and Tab focus cycling between panels.

### cmd/hexedit -- Hex Editor

A full hex editor with offset/hex/ASCII columns, a data inspector panel showing
multi-width integer interpretations, search by string/hex/integer with alignment
constraints, and goto-offset dialog. Supports keyboard navigation, mouse
clicking in hex and ASCII regions, and mouse wheel scrolling.

### cmd/dungeon -- Dungeon Explorer

An Angband-style procedural cave exploration game using cellular automata
generation. Features fog of war with circular line-of-sight, auto-scrolling
camera, and multiple tile types (floor, wall, water, doors, stairs). Uses the
Surface virtual canvas for rendering the 200x200 world.

### cmd/menu -- Menu Bar Demo

A horizontal menu bar with dropdown overlay menus. Demonstrates the overlay
system with z-ordered floating panels, Alt+letter accelerators, full keyboard
navigation (arrows, Enter, Escape), and mouse click support on both the bar
and dropdown items.

### cmd/fsview -- Filesystem Viewer

A tree-based filesystem browser with a preview panel. The left pane shows a
navigable directory tree with lazy loading and expand/collapse. The right pane
shows file metadata, text content preview with line numbers, or hex dump for
binary files.

## Building

```
go build ./cmd/...
```

Run individual applications:

```
go run ./cmd/demo
go run ./cmd/hexedit <file>
go run ./cmd/dungeon
go run ./cmd/menu
go run ./cmd/fsview [path]
```

## Requirements

- Go 1.25+
- Linux terminal with ANSI escape sequence support
- Terminal must support SGR mouse mode (1003+1006) for mouse interaction
