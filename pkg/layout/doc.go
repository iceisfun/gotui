// Package layout provides the application shell and layout containers for
// building terminal UI applications.
//
// [App] is the top-level event loop owner. It manages the terminal lifecycle,
// input decoding, event dispatch (keyboard, mouse, resize), focus management
// with Tab/Backtab cycling, global key bindings, periodic tick updates for
// animation, and multi-layer compositing with overlay support.
//
// Layout containers divide screen space among child renderables:
//   - [Split] arranges children along a horizontal or vertical axis with
//     proportional or fixed sizing.
//   - [Absolute] positions children at explicit coordinates.
//   - [Viewport] wraps a child in a scrollable window with mouse wheel support.
//   - [Modal] is a centered floating overlay that captures all input when visible.
//
// The overlay system collects [render.OverlayRequest] values from the renderable
// tree, renders them into a separate transparent buffer layer, and composites
// them above the base content. Mouse hit-testing walks overlays first (highest
// z-order to lowest), then the container tree.
package layout
