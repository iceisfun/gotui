# cmd/fsview -- Filesystem Viewer

A tree-based filesystem browser with a split-pane layout. The left panel
shows a navigable directory tree with lazy-loaded children and
expand/collapse support. The right panel displays a preview of the selected
entry: file metadata, text content with line numbers, or a hex dump for
binary files.

## How to Run

```
go run ./cmd/fsview [path]
```

If no path is given, the current directory is used.

## Controls

| Key | Action |
|-----|--------|
| Up / Down | Move cursor in the tree |
| Right / Enter | Expand directory or select file |
| Left | Collapse directory |
| h | Toggle hidden file visibility |
| r | Refresh the tree from disk |
| Tab | Cycle focus between tree and preview panels |
| q | Quit |
| Mouse click | Select tree node |
| Mouse wheel | Scroll the tree |
