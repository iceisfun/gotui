package input

import "strconv"

// parseSGRMouse parses an SGR mouse sequence. The input is the content between
// ESC[< and the final M or m character. For example, for ESC[<0;10;20M the
// input would be "0;10;20" and final='M'.
func parseSGRMouse(params string, final byte) (MouseEvent, bool) {
	// Format: ESC[<btn;x;y{M|m}
	// btn encodes button number + modifier bits.
	// M = press, m = release.
	parts := splitSemicolon(params)
	if len(parts) != 3 {
		return MouseEvent{}, false
	}

	btn, err := strconv.Atoi(parts[0])
	if err != nil {
		return MouseEvent{}, false
	}
	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return MouseEvent{}, false
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return MouseEvent{}, false
	}

	// Convert from 1-based to 0-based coordinates.
	x--
	y--

	ev := MouseEvent{X: x, Y: y, ScreenX: x, ScreenY: y}

	// Extract modifier bits from btn.
	if btn&4 != 0 {
		ev.Mod |= ModShift
	}
	if btn&8 != 0 {
		ev.Mod |= ModAlt
	}
	if btn&16 != 0 {
		ev.Mod |= ModCtrl
	}

	// Button encoding (lower 2 bits + bit 6).
	buttonCode := (btn & 3) | ((btn >> 4) & 0x4)
	motion := btn&32 != 0

	if final == 'm' {
		ev.Button = MouseRelease
	} else if motion {
		ev.Button = MouseMotion
	} else {
		switch buttonCode {
		case 0:
			ev.Button = MouseLeft
		case 1:
			ev.Button = MouseMiddle
		case 2:
			ev.Button = MouseRight
		case 4:
			ev.Button = MouseWheelUp
		case 5:
			ev.Button = MouseWheelDown
		default:
			ev.Button = MouseNone
		}
	}

	return ev, true
}

func splitSemicolon(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ';' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
