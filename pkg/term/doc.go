// Package term manages raw terminal I/O for full-screen TUI applications.
//
// [Terminal] puts the controlling terminal into raw mode, enables mouse tracking
// (SGR mode with all-motion reporting), bracketed paste, and the alternate screen
// buffer. Call [Open] to enter raw mode and Close to restore the original state.
//
// [Writer] is a buffered ANSI escape sequence builder that accumulates cursor
// movement, style changes, and text output, then flushes them in a single write
// to minimize flickering. [Reader] reads raw bytes from the terminal in a
// background goroutine, delivering chunks via a channel.
package term
