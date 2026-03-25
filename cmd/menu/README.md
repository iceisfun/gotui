# cmd/menu -- Menu Bar Demo

Demonstrates a horizontal menu bar with dropdown overlay menus. The bar
contains File, Edit, View, and Help menus. Selecting a menu item logs the
action in the scrollable content area below.

This demo exercises the overlay system: dropdown menus are rendered as
transparent overlay layers composited above the base content, with proper
mouse hit-testing that dispatches clicks to the overlay before the content
beneath.

## How to Run

```
go run ./cmd/menu
```

## Controls

| Key | Action |
|-----|--------|
| Alt+F | Open/close File menu |
| Alt+E | Open/close Edit menu |
| Alt+V | Open/close View menu |
| Alt+H | Open/close Help menu |
| Up / Down | Navigate items within an open menu |
| Left / Right | Switch to adjacent menu (while a menu is open) |
| Enter | Select the highlighted menu item |
| Escape | Close the open menu |
| Ctrl-C | Quit |
| Mouse click on bar | Open/close the clicked menu |
| Mouse click on item | Select the clicked menu item |
