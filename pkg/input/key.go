package input

// ModMask represents keyboard modifier keys as a bitmask.
type ModMask uint8

const (
	ModShift ModMask = 1 << iota
	ModAlt
	ModCtrl
)

// KeyCode identifies a key on the keyboard.
type KeyCode uint16

const (
	KeyNone KeyCode = iota
	KeyRune         // Printable character — use KeyEvent.Rune.
	KeyEnter
	KeyTab
	KeyBacktab // Shift+Tab
	KeyBackspace
	KeyEscape
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyDelete
	KeyInsert
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// String returns a human-readable name for the key code.
func (k KeyCode) String() string {
	switch k {
	case KeyNone:
		return "None"
	case KeyRune:
		return "Rune"
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyBacktab:
		return "Backtab"
	case KeyBackspace:
		return "Backspace"
	case KeyEscape:
		return "Escape"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPgUp:
		return "PgUp"
	case KeyPgDown:
		return "PgDown"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	}
	if k >= KeyF1 && k <= KeyF12 {
		return "F" + string(rune('0'+int(k-KeyF1+1)))
	}
	return "Unknown"
}

// String returns a human-readable name for modifier flags.
func (m ModMask) String() string {
	s := ""
	if m&ModCtrl != 0 {
		s += "Ctrl+"
	}
	if m&ModAlt != 0 {
		s += "Alt+"
	}
	if m&ModShift != 0 {
		s += "Shift+"
	}
	return s
}
