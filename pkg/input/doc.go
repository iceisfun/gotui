// Package input decodes raw terminal byte sequences into structured input events.
//
// The [Decoder] is a state machine that processes bytes from the terminal reader
// and produces [Event] values representing keyboard input, mouse actions, terminal
// resize notifications, and bracketed paste content.
//
// Keyboard events are described by [KeyCode] (special keys like arrows, function
// keys, Enter, Escape) and [ModMask] (Ctrl, Alt, Shift). Printable characters
// use [KeyRune] with the rune value in [KeyEvent].Rune.
//
// Mouse events use SGR encoding (mode 1003+1006) and report button presses,
// releases, wheel scroll, and motion with both local and absolute screen
// coordinates in [MouseEvent].
package input
