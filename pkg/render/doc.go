// Package render provides the core rendering engine for terminal UI applications.
//
// The rendering model is cell-based: a [Buffer] is a 2D grid of [Cell] values,
// where each cell holds a rune, display width, style, and a transparency flag.
// [View] provides a clipped, offset window into a buffer for safe sub-region
// rendering with automatic bounds checking.
//
// Compositing is handled by [Composite] and [CompositeTwo], which flatten a
// stack of buffer layers (bottom-to-top) into a destination buffer. Transparent
// cells fall through to the layer below, enabling overlay effects.
//
// [Flush] performs differential updates, writing only changed cells to the
// terminal via ANSI escape sequences for efficient screen refresh.
//
// The [Renderable] interface is the foundation of the component model.
// Additional interfaces ([Interactive], [Focusable], [Scrollable], [Overlayable],
// [Cursorable], [Container], [MouseTarget]) extend components with input handling,
// focus management, scrolling, overlay support, and layout participation.
//
// [Surface] implements an infinite scrollable virtual canvas with a camera
// system, suitable for applications like map viewers and dungeon explorers.
package render
