package input

import "unicode/utf8"

type decodeState uint8

const (
	stateGround decodeState = iota
	stateEsc                // Saw ESC, waiting for [ or O or character
	stateCSI                // Inside ESC[ CSI sequence
	stateCSIParam           // Accumulating CSI params
	stateSS3                // Inside ESC O SS3 sequence
	statePaste              // Inside bracketed paste ESC[200~
	stateSGRMouse           // Inside SGR mouse ESC[<
)

// Decoder transforms raw terminal bytes into structured Events.
type Decoder struct {
	buf       []byte
	state     decodeState
	params    []byte // CSI parameter accumulator
	pasteData []byte // paste content accumulator
}

// NewDecoder creates a new input decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Feed processes raw bytes and returns decoded events.
func (d *Decoder) Feed(raw []byte) []Event {
	d.buf = append(d.buf, raw...)
	var events []Event

	for len(d.buf) > 0 {
		ev, consumed := d.decode()
		if consumed == 0 {
			break
		}
		d.buf = d.buf[consumed:]
		if ev != nil {
			events = append(events, *ev)
		}
	}

	return events
}

func (d *Decoder) decode() (*Event, int) {
	switch d.state {
	case stateGround:
		return d.decodeGround()
	case stateEsc:
		return d.decodeEsc()
	case stateCSI, stateCSIParam:
		return d.decodeCSI()
	case stateSS3:
		return d.decodeSS3()
	case statePaste:
		return d.decodePaste()
	case stateSGRMouse:
		return d.decodeSGRMouse()
	}
	// Should not happen; reset.
	d.state = stateGround
	return nil, 1
}

func (d *Decoder) decodeGround() (*Event, int) {
	b := d.buf[0]

	// ESC
	if b == 0x1b {
		if len(d.buf) == 1 {
			// Might be start of escape sequence or lone Escape.
			// Since we coalesce bytes, if there's only one byte it's a lone ESC.
			ev := keyEvent(KeyEscape, 0, 0)
			return &ev, 1
		}
		d.state = stateEsc
		return nil, 1
	}

	// Ctrl+letter (0x01–0x1a except tab, enter, backspace)
	if b >= 0x01 && b <= 0x1a {
		switch b {
		case 0x09: // Tab
			ev := keyEvent(KeyTab, 0, 0)
			return &ev, 1
		case 0x0d: // Enter (CR)
			ev := keyEvent(KeyEnter, 0, 0)
			return &ev, 1
		case 0x08: // Backspace (some terminals)
			ev := keyEvent(KeyBackspace, 0, 0)
			return &ev, 1
		default:
			// Ctrl+A through Ctrl+Z
			r := rune('a' + b - 1)
			ev := keyEvent(KeyRune, ModCtrl, r)
			return &ev, 1
		}
	}

	// DEL (backspace on most terminals)
	if b == 0x7f {
		ev := keyEvent(KeyBackspace, 0, 0)
		return &ev, 1
	}

	// Regular UTF-8 character.
	r, size := utf8.DecodeRune(d.buf)
	if r == utf8.RuneError && size <= 1 {
		if len(d.buf) < 4 {
			// Might be incomplete UTF-8; wait for more bytes.
			// But since we coalesce, treat as invalid.
			return nil, 1
		}
		return nil, 1
	}
	ev := keyEvent(KeyRune, 0, r)
	return &ev, size
}

func (d *Decoder) decodeEsc() (*Event, int) {
	if len(d.buf) == 0 {
		// ESC was consumed, nothing follows — emit Escape.
		d.state = stateGround
		ev := keyEvent(KeyEscape, 0, 0)
		return &ev, 0
	}

	b := d.buf[0]

	switch b {
	case '[':
		d.state = stateCSI
		d.params = d.params[:0]
		return nil, 1
	case 'O':
		d.state = stateSS3
		return nil, 1
	default:
		// Alt+key: ESC followed by a printable character.
		d.state = stateGround
		if b >= 0x20 && b < 0x7f {
			ev := keyEvent(KeyRune, ModAlt, rune(b))
			return &ev, 1
		}
		// Alt+Ctrl
		if b >= 0x01 && b <= 0x1a {
			r := rune('a' + b - 1)
			ev := keyEvent(KeyRune, ModAlt|ModCtrl, r)
			return &ev, 1
		}
		ev := keyEvent(KeyEscape, 0, 0)
		return &ev, 0
	}
}

func (d *Decoder) decodeCSI() (*Event, int) {
	if len(d.buf) == 0 {
		return nil, 0
	}

	b := d.buf[0]

	// Check for SGR mouse: ESC[<
	if len(d.params) == 0 && b == '<' {
		d.state = stateSGRMouse
		d.params = d.params[:0]
		return nil, 1
	}

	// Check for bracketed paste start: ESC[200~
	if len(d.params) == 0 && b >= '0' && b <= '9' {
		d.params = append(d.params, b)
		d.state = stateCSIParam
		return nil, 1
	}

	// Parameter bytes: digits (0x30–0x39) and semicolons (0x3B).
	// Anything in 0x40–0x7E is a final byte that terminates the sequence.
	if (b >= '0' && b <= '9') || b == ';' {
		d.params = append(d.params, b)
		d.state = stateCSIParam
		if len(d.params) > 32 {
			d.state = stateGround
			d.params = d.params[:0]
		}
		return nil, 1
	}

	// Final byte — decode the sequence.
	d.state = stateGround
	ev := d.interpretCSI(string(d.params), b)
	d.params = d.params[:0]
	if ev != nil {
		return ev, 1
	}
	return nil, 1
}

func (d *Decoder) interpretCSI(params string, final byte) *Event {
	// Check for bracketed paste.
	if final == '~' {
		switch params {
		case "200":
			d.state = statePaste
			d.pasteData = d.pasteData[:0]
			return nil
		}
	}

	// Tilde-terminated sequences: ESC[N~
	if final == '~' {
		mod := ModMask(0)
		code := params
		if semi := indexByte(params, ';'); semi >= 0 {
			code = params[:semi]
			mod = parseCSIMod(params[semi+1:])
		}
		switch code {
		case "1", "7":
			ev := keyEvent(KeyHome, mod, 0)
			return &ev
		case "2":
			ev := keyEvent(KeyInsert, mod, 0)
			return &ev
		case "3":
			ev := keyEvent(KeyDelete, mod, 0)
			return &ev
		case "4", "8":
			ev := keyEvent(KeyEnd, mod, 0)
			return &ev
		case "5":
			ev := keyEvent(KeyPgUp, mod, 0)
			return &ev
		case "6":
			ev := keyEvent(KeyPgDown, mod, 0)
			return &ev
		case "11":
			ev := keyEvent(KeyF1, mod, 0)
			return &ev
		case "12":
			ev := keyEvent(KeyF2, mod, 0)
			return &ev
		case "13":
			ev := keyEvent(KeyF3, mod, 0)
			return &ev
		case "14":
			ev := keyEvent(KeyF4, mod, 0)
			return &ev
		case "15":
			ev := keyEvent(KeyF5, mod, 0)
			return &ev
		case "17":
			ev := keyEvent(KeyF6, mod, 0)
			return &ev
		case "18":
			ev := keyEvent(KeyF7, mod, 0)
			return &ev
		case "19":
			ev := keyEvent(KeyF8, mod, 0)
			return &ev
		case "20":
			ev := keyEvent(KeyF9, mod, 0)
			return &ev
		case "21":
			ev := keyEvent(KeyF10, mod, 0)
			return &ev
		case "23":
			ev := keyEvent(KeyF11, mod, 0)
			return &ev
		case "24":
			ev := keyEvent(KeyF12, mod, 0)
			return &ev
		}
		return nil
	}

	// Arrow keys and Home/End: ESC[A, ESC[1;modA, etc.
	mod := ModMask(0)
	if semi := indexByte(params, ';'); semi >= 0 {
		mod = parseCSIMod(params[semi+1:])
	}

	switch final {
	case 'A':
		ev := keyEvent(KeyUp, mod, 0)
		return &ev
	case 'B':
		ev := keyEvent(KeyDown, mod, 0)
		return &ev
	case 'C':
		ev := keyEvent(KeyRight, mod, 0)
		return &ev
	case 'D':
		ev := keyEvent(KeyLeft, mod, 0)
		return &ev
	case 'H':
		ev := keyEvent(KeyHome, mod, 0)
		return &ev
	case 'F':
		ev := keyEvent(KeyEnd, mod, 0)
		return &ev
	case 'Z':
		ev := keyEvent(KeyBacktab, 0, 0)
		return &ev
	}

	return nil
}

func (d *Decoder) decodeSS3() (*Event, int) {
	if len(d.buf) == 0 {
		return nil, 0
	}

	d.state = stateGround
	b := d.buf[0]

	switch b {
	case 'A':
		ev := keyEvent(KeyUp, 0, 0)
		return &ev, 1
	case 'B':
		ev := keyEvent(KeyDown, 0, 0)
		return &ev, 1
	case 'C':
		ev := keyEvent(KeyRight, 0, 0)
		return &ev, 1
	case 'D':
		ev := keyEvent(KeyLeft, 0, 0)
		return &ev, 1
	case 'H':
		ev := keyEvent(KeyHome, 0, 0)
		return &ev, 1
	case 'F':
		ev := keyEvent(KeyEnd, 0, 0)
		return &ev, 1
	case 'P':
		ev := keyEvent(KeyF1, 0, 0)
		return &ev, 1
	case 'Q':
		ev := keyEvent(KeyF2, 0, 0)
		return &ev, 1
	case 'R':
		ev := keyEvent(KeyF3, 0, 0)
		return &ev, 1
	case 'S':
		ev := keyEvent(KeyF4, 0, 0)
		return &ev, 1
	}

	return nil, 1
}

func (d *Decoder) decodePaste() (*Event, int) {
	// Look for paste end sequence: ESC[201~
	// We scan for 0x1b in the buffer.
	for i := 0; i < len(d.buf); i++ {
		if d.buf[i] == 0x1b {
			// Check for ESC[201~
			end := "\x1b[201~"
			if i+len(end) <= len(d.buf) && string(d.buf[i:i+len(end)]) == end {
				d.pasteData = append(d.pasteData, d.buf[:i]...)
				d.state = stateGround
				ev := Event{
					Type:  EventPaste,
					Paste: string(d.pasteData),
				}
				d.pasteData = d.pasteData[:0]
				return &ev, i + len(end)
			}
		}
	}

	// No end marker found yet; accumulate everything.
	d.pasteData = append(d.pasteData, d.buf...)
	return nil, len(d.buf)
}

func (d *Decoder) decodeSGRMouse() (*Event, int) {
	// Accumulate until M or m.
	for i := 0; i < len(d.buf); i++ {
		b := d.buf[i]
		if b == 'M' || b == 'm' {
			params := string(d.params)
			d.params = d.params[:0]
			d.state = stateGround
			if me, ok := parseSGRMouse(params, b); ok {
				ev := Event{Type: EventMouse, Mouse: me}
				return &ev, i + 1
			}
			return nil, i + 1
		}
		d.params = append(d.params, b)
		if len(d.params) > 32 {
			d.state = stateGround
			d.params = d.params[:0]
			return nil, i + 1
		}
	}
	return nil, 0 // Need more data.
}

// parseCSIMod converts the modifier parameter from xterm-style CSI sequences.
// The modifier is encoded as (mod_value + 1), where mod_value has bits:
// 1=Shift, 2=Alt, 4=Ctrl.
func parseCSIMod(s string) ModMask {
	n := 0
	for _, b := range []byte(s) {
		if b >= '0' && b <= '9' {
			n = n*10 + int(b-'0')
		}
	}
	if n <= 1 {
		return 0
	}
	n-- // xterm encodes mod+1
	mod := ModMask(0)
	if n&1 != 0 {
		mod |= ModShift
	}
	if n&2 != 0 {
		mod |= ModAlt
	}
	if n&4 != 0 {
		mod |= ModCtrl
	}
	return mod
}

func keyEvent(code KeyCode, mod ModMask, r rune) Event {
	return Event{
		Type: EventKey,
		Key: KeyEvent{
			Code: code,
			Mod:  mod,
			Rune: r,
		},
	}
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
