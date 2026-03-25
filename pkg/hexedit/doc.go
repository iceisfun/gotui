// Package hexedit provides components for building a hex editor application.
//
// [Buffer] loads a file into memory and provides byte-level access, numeric
// interpretation at any offset ([Interpretation]), and byte pattern search with
// optional alignment constraints.
//
// [HexView] is the main renderable that displays offset, hex, and ASCII columns
// with cursor navigation, search result highlighting, and mouse click hit-testing
// in both hex and ASCII regions.
//
// [InfoPanel] displays a data inspector showing signed/unsigned integer
// interpretations (8/16/32/64-bit, little-endian and big-endian), binary
// representation, and hex dump at the cursor position.
//
// [SearchDialog] and [GotoDialog] are modal input dialogs for searching by
// string, hex bytes, or integer values (with alignment options), and for
// jumping to a specific byte offset (decimal, hex, octal, or relative).
package hexedit
