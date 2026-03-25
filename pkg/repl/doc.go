// Package repl provides a full-featured interactive read-eval-print loop component.
//
// The [REPL] type is a renderable widget that implements Focusable, Cursorable,
// Scrollable, and Overlayable. It combines a multiline [Editor] with command
// [History], a [CompletionPopup] overlay, syntax highlighting, and an [Output]
// display area.
//
// The host application provides an [Executor] to evaluate user input and an
// optional [Completer] for tab-completion proposals and [Highlighter] for live
// syntax coloring. The REPL itself has no dependency on any particular language.
//
// Key bindings follow readline conventions: Ctrl-A/E for home/end, Ctrl-K/U for
// line kill, Ctrl-W for word delete, Ctrl-P/N or Up/Down for history navigation,
// and Tab for completion.
package repl
