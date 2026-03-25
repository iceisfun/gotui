# cmd/hexedit -- Hex Editor

A terminal hex editor with a data inspector panel, search with multiple modes
and alignment options, and goto-offset dialog.

The main view displays three columns: offset (hex), hex byte values (grouped
in pairs of 8), and ASCII representation. The right panel is an inspector
showing signed and unsigned integer interpretations at the cursor position
in 8, 16, 32, and 64-bit widths (both little-endian and big-endian), plus
binary and hex representations.

## How to Run

```
go run ./cmd/hexedit <file>
```

## Key Bindings

| Key | Action |
|-----|--------|
| Arrow keys | Move cursor by one byte (left/right) or one row (up/down) |
| PgUp / PgDown | Scroll by one visible page |
| Home | Move to start of current row (Ctrl-Home: start of file) |
| End | Move to end of current row (Ctrl-End: end of file) |
| Ctrl-F or / | Open search dialog |
| Ctrl-N | Find next occurrence of last search |
| Ctrl-G or Ctrl-J | Open goto-offset dialog |
| Ctrl-C | Quit |
| Mouse click | Click on hex or ASCII column to position cursor |
| Mouse wheel | Scroll up/down |

## Search Dialog

Press Ctrl-F to open. The dialog supports multiple search modes:

| Mode | Input format |
|------|-------------|
| string | Plain text (UTF-8 bytes) |
| bytes | Hex bytes, e.g. `DE AD BE EF` or `DEADBEEF` |
| u8 | 8-bit unsigned integer |
| u16 | 16-bit unsigned integer (searched as little-endian) |
| u32 | 32-bit unsigned integer (searched as little-endian) |
| u64 | 64-bit unsigned integer (searched as little-endian) |

Within the search dialog:

- **Tab / Shift-Tab**: Cycle search mode
- **Alt-A**: Cycle alignment constraint (Unaligned, 32-bit, 64-bit)
- **Enter**: Execute search
- **Escape**: Cancel

## Goto Dialog

Press Ctrl-G to open. Accepts:

- Decimal offset: `12345`
- Hex offset: `0x3039`
- Octal offset: `0o30071`
- Relative offset: `+100` or `-50` (from current cursor position)
