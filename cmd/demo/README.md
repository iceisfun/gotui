# cmd/demo -- Click Hit-Test Demo

Demonstrates mouse hit-testing through the layout tree, overlay compositing,
and focus cycling between panels.

The screen is split into two panels, each containing colored clickable buttons.
Clicking anywhere in a panel logs the local and screen coordinates, verifying
that mouse events are correctly translated through the container hierarchy.
A floating overlay panel can be toggled to test overlay hit-testing and
compositing with transparency.

## How to Run

```
go run ./cmd/demo
```

## Key Bindings

| Key | Action |
|-----|--------|
| Tab | Cycle focus between left and right panels |
| Ctrl-O | Toggle floating overlay panel |
| Escape | Close overlay (if open) |
| Ctrl-C | Quit |
| Mouse click | Click buttons or anywhere in a panel to see coordinate logging |
