// Package widget provides ready-to-use UI components for terminal applications.
//
// Widgets implement combinations of the render package interfaces (Renderable,
// Interactive, Focusable, Scrollable, Container) and can be composed into
// layout trees.
//
// Available widgets:
//   - [Label]: static styled text display with left, center, or right alignment.
//   - [Button]: clickable button activated by Enter or mouse click.
//   - [Selector]: horizontal option picker navigated with Left/Right arrows or mouse.
//   - [List]: vertical scrollable list with keyboard and mouse selection.
//   - [Tree]: expandable/collapsible tree view with lazy node expansion.
//   - [Menu]: popup dropdown menu with keyboard navigation and mouse support.
//   - [Panel]: bordered container with an optional title and single, double,
//     or rounded border styles.
//   - [TextView]: scrollable multi-line styled text display with auto-scroll.
package widget
