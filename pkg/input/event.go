package input

import "fmt"

// EventType classifies an input event.
type EventType uint8

const (
	EventKey    EventType = iota // Keyboard input
	EventMouse                   // Mouse input
	EventResize                  // Terminal resize
	EventPaste                   // Bracketed paste content
)

// Event is a decoded terminal input event.
type Event struct {
	Type  EventType
	Key   KeyEvent
	Mouse MouseEvent
	Size  SizeEvent
	Paste string
}

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	Code KeyCode
	Mod  ModMask
	Rune rune // Valid when Code == KeyRune.
}

// String returns a human-readable description of the key event.
func (k KeyEvent) String() string {
	mod := k.Mod.String()
	if k.Code == KeyRune {
		return fmt.Sprintf("%s%c", mod, k.Rune)
	}
	return mod + k.Code.String()
}

// MouseButton identifies a mouse button.
type MouseButton uint8

const (
	MouseNone       MouseButton = iota
	MouseLeft                   // Button 1
	MouseMiddle                 // Button 2
	MouseRight                  // Button 3
	MouseRelease                // Button release
	MouseWheelUp                // Scroll up
	MouseWheelDown              // Scroll down
	MouseMotion                 // Motion with no button change
)

// MouseEvent represents a mouse event.
// X, Y are local coordinates relative to the receiving renderable.
// ScreenX, ScreenY are absolute terminal coordinates (always set by the decoder).
// The layout system translates X, Y as events flow through the tree.
type MouseEvent struct {
	X, Y         int // Local to the receiving element.
	ScreenX, ScreenY int // Absolute terminal coordinates.
	Button MouseButton
	Mod    ModMask
}

// SizeEvent represents a terminal resize event.
type SizeEvent struct {
	Cols int
	Rows int
}
