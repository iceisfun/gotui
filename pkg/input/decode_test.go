package input

import "testing"

func feed(raw ...byte) []Event {
	d := NewDecoder()
	return d.Feed(raw)
}

func requireOneKey(t *testing.T, events []Event) KeyEvent {
	t.Helper()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Type != EventKey {
		t.Fatalf("expected EventKey, got %v", events[0].Type)
	}
	return events[0].Key
}

func TestRegularASCII(t *testing.T) {
	tests := []struct {
		char byte
		want rune
	}{
		{'a', 'a'},
		{'z', 'z'},
		{'A', 'A'},
		{'0', '0'},
		{' ', ' '},
		{'~', '~'},
	}
	for _, tt := range tests {
		k := requireOneKey(t, feed(tt.char))
		if k.Code != KeyRune || k.Rune != tt.want {
			t.Fatalf("byte %d: got Code=%v Rune=%c, want KeyRune %c", tt.char, k.Code, k.Rune, tt.want)
		}
		if k.Mod != 0 {
			t.Fatalf("byte %d: unexpected mod %v", tt.char, k.Mod)
		}
	}
}

func TestUTF8MultiByte(t *testing.T) {
	// é is 0xc3 0xa9 in UTF-8
	events := feed(0xc3, 0xa9)
	k := requireOneKey(t, events)
	if k.Code != KeyRune || k.Rune != 'é' {
		t.Fatalf("got Code=%v Rune=%c, want é", k.Code, k.Rune)
	}
}

func TestUTF8ThreeByte(t *testing.T) {
	// 世 is 0xe4 0xb8 0x96
	events := feed(0xe4, 0xb8, 0x96)
	k := requireOneKey(t, events)
	if k.Code != KeyRune || k.Rune != '世' {
		t.Fatalf("got %c, want 世", k.Rune)
	}
}

func TestCtrlLetters(t *testing.T) {
	tests := []struct {
		byte byte
		rune rune
	}{
		{0x01, 'a'}, // Ctrl+A
		{0x03, 'c'}, // Ctrl+C
		{0x1a, 'z'}, // Ctrl+Z
	}
	for _, tt := range tests {
		k := requireOneKey(t, feed(tt.byte))
		if k.Code != KeyRune || k.Mod != ModCtrl || k.Rune != tt.rune {
			t.Fatalf("byte 0x%02x: got Code=%v Mod=%v Rune=%c, want Ctrl+%c",
				tt.byte, k.Code, k.Mod, k.Rune, tt.rune)
		}
	}
}

func TestEscapeKey(t *testing.T) {
	// Lone ESC byte.
	k := requireOneKey(t, feed(0x1b))
	if k.Code != KeyEscape {
		t.Fatalf("got Code=%v, want KeyEscape", k.Code)
	}
}

func TestAltLetter(t *testing.T) {
	// ESC + 'a'
	events := feed(0x1b, 'a')
	k := requireOneKey(t, events)
	if k.Code != KeyRune || k.Mod != ModAlt || k.Rune != 'a' {
		t.Fatalf("got Code=%v Mod=%v Rune=%c, want Alt+a", k.Code, k.Mod, k.Rune)
	}
}

func TestArrowKeys(t *testing.T) {
	tests := []struct {
		name   string
		seq    []byte
		expect KeyCode
	}{
		{"Up", []byte{0x1b, '[', 'A'}, KeyUp},
		{"Down", []byte{0x1b, '[', 'B'}, KeyDown},
		{"Right", []byte{0x1b, '[', 'C'}, KeyRight},
		{"Left", []byte{0x1b, '[', 'D'}, KeyLeft},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := requireOneKey(t, feed(tt.seq...))
			if k.Code != tt.expect {
				t.Fatalf("got %v, want %v", k.Code, tt.expect)
			}
			if k.Mod != 0 {
				t.Fatalf("unexpected mod: %v", k.Mod)
			}
		})
	}
}

func TestArrowKeysWithModifiers(t *testing.T) {
	tests := []struct {
		name string
		seq  []byte
		code KeyCode
		mod  ModMask
	}{
		// ESC[1;5A = Ctrl+Up (5 = 4+1, 4 = Ctrl)
		{"Ctrl+Up", []byte{0x1b, '[', '1', ';', '5', 'A'}, KeyUp, ModCtrl},
		// ESC[1;2B = Shift+Down
		{"Shift+Down", []byte{0x1b, '[', '1', ';', '2', 'B'}, KeyDown, ModShift},
		// ESC[1;3C = Alt+Right
		{"Alt+Right", []byte{0x1b, '[', '1', ';', '3', 'C'}, KeyRight, ModAlt},
		// ESC[1;8D = Ctrl+Alt+Shift+Left (8=7+1, 7=1+2+4)
		{"Ctrl+Alt+Shift+Left", []byte{0x1b, '[', '1', ';', '8', 'D'}, KeyLeft, ModShift | ModAlt | ModCtrl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := requireOneKey(t, feed(tt.seq...))
			if k.Code != tt.code {
				t.Fatalf("got %v, want %v", k.Code, tt.code)
			}
			if k.Mod != tt.mod {
				t.Fatalf("mod = %v, want %v", k.Mod, tt.mod)
			}
		})
	}
}

func TestFunctionKeysTilde(t *testing.T) {
	tests := []struct {
		name   string
		params string
		expect KeyCode
	}{
		{"F5", "15", KeyF5},
		{"F6", "17", KeyF6},
		{"F7", "18", KeyF7},
		{"F8", "19", KeyF8},
		{"F9", "20", KeyF9},
		{"F10", "21", KeyF10},
		{"F11", "23", KeyF11},
		{"F12", "24", KeyF12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := append([]byte{0x1b, '['}, []byte(tt.params)...)
			seq = append(seq, '~')
			k := requireOneKey(t, feed(seq...))
			if k.Code != tt.expect {
				t.Fatalf("got %v, want %v", k.Code, tt.expect)
			}
		})
	}
}

func TestFunctionKeysSS3(t *testing.T) {
	tests := []struct {
		name   string
		final  byte
		expect KeyCode
	}{
		{"F1", 'P', KeyF1},
		{"F2", 'Q', KeyF2},
		{"F3", 'R', KeyF3},
		{"F4", 'S', KeyF4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := requireOneKey(t, feed(0x1b, 'O', tt.final))
			if k.Code != tt.expect {
				t.Fatalf("got %v, want %v", k.Code, tt.expect)
			}
		})
	}
}

func TestHomeEnd(t *testing.T) {
	tests := []struct {
		name string
		seq  []byte
		code KeyCode
	}{
		{"Home CSI", []byte{0x1b, '[', 'H'}, KeyHome},
		{"End CSI", []byte{0x1b, '[', 'F'}, KeyEnd},
		{"Home SS3", []byte{0x1b, 'O', 'H'}, KeyHome},
		{"End SS3", []byte{0x1b, 'O', 'F'}, KeyEnd},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := requireOneKey(t, feed(tt.seq...))
			if k.Code != tt.code {
				t.Fatalf("got %v, want %v", k.Code, tt.code)
			}
		})
	}
}

func TestPgUpPgDown(t *testing.T) {
	// ESC[5~ = PgUp, ESC[6~ = PgDown
	k := requireOneKey(t, feed(0x1b, '[', '5', '~'))
	if k.Code != KeyPgUp {
		t.Fatalf("got %v, want PgUp", k.Code)
	}
	k = requireOneKey(t, feed(0x1b, '[', '6', '~'))
	if k.Code != KeyPgDown {
		t.Fatalf("got %v, want PgDown", k.Code)
	}
}

func TestDelete(t *testing.T) {
	k := requireOneKey(t, feed(0x1b, '[', '3', '~'))
	if k.Code != KeyDelete {
		t.Fatalf("got %v, want Delete", k.Code)
	}
}

func TestTabEnterBackspace(t *testing.T) {
	k := requireOneKey(t, feed(0x09))
	if k.Code != KeyTab {
		t.Fatalf("got %v, want Tab", k.Code)
	}
	k = requireOneKey(t, feed(0x0d))
	if k.Code != KeyEnter {
		t.Fatalf("got %v, want Enter", k.Code)
	}
	k = requireOneKey(t, feed(0x7f))
	if k.Code != KeyBackspace {
		t.Fatalf("got %v, want Backspace", k.Code)
	}
}

func TestBacktab(t *testing.T) {
	// ESC[Z = Backtab (Shift+Tab)
	k := requireOneKey(t, feed(0x1b, '[', 'Z'))
	if k.Code != KeyBacktab {
		t.Fatalf("got %v, want Backtab", k.Code)
	}
}

func TestSGRMousePress(t *testing.T) {
	// ESC[<0;10;20M = left button press at (10,20)
	seq := []byte("\x1b[<0;10;20M")
	d := NewDecoder()
	events := d.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventMouse {
		t.Fatalf("expected EventMouse, got %v", events[0].Type)
	}
	m := events[0].Mouse
	if m.Button != MouseLeft {
		t.Fatalf("button = %v, want MouseLeft", m.Button)
	}
	if m.X != 9 || m.Y != 19 {
		t.Fatalf("coords = (%d,%d), want (9,19)", m.X, m.Y)
	}
}

func TestSGRMouseRelease(t *testing.T) {
	// ESC[<0;10;20m = release (lowercase m)
	seq := []byte("\x1b[<0;10;20m")
	d := NewDecoder()
	events := d.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	m := events[0].Mouse
	if m.Button != MouseRelease {
		t.Fatalf("button = %v, want MouseRelease", m.Button)
	}
}

func TestMouseWheel(t *testing.T) {
	tests := []struct {
		name   string
		seq    string
		button MouseButton
	}{
		{"WheelUp", "\x1b[<64;10;20M", MouseWheelUp},
		{"WheelDown", "\x1b[<65;10;20M", MouseWheelDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder()
			events := d.Feed([]byte(tt.seq))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Mouse.Button != tt.button {
				t.Fatalf("button = %v, want %v", events[0].Mouse.Button, tt.button)
			}
		})
	}
}

func TestBracketedPaste(t *testing.T) {
	// ESC[200~ content ESC[201~
	seq := []byte("\x1b[200~hello world\x1b[201~")
	d := NewDecoder()
	events := d.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Type != EventPaste {
		t.Fatalf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != "hello world" {
		t.Fatalf("Paste = %q, want %q", events[0].Paste, "hello world")
	}
}

func TestMultipleEventsInOneFeed(t *testing.T) {
	// Feed two regular characters at once.
	d := NewDecoder()
	events := d.Feed([]byte("ab"))
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Key.Rune != 'a' || events[1].Key.Rune != 'b' {
		t.Fatal("wrong runes")
	}
}

func TestInsertKey(t *testing.T) {
	// ESC[2~
	k := requireOneKey(t, feed(0x1b, '[', '2', '~'))
	if k.Code != KeyInsert {
		t.Fatalf("got %v, want Insert", k.Code)
	}
}

func TestF1ThroughF4Tilde(t *testing.T) {
	// Some terminals send F1-F4 as ESC[11~ through ESC[14~
	tests := []struct {
		params string
		code   KeyCode
	}{
		{"11", KeyF1},
		{"12", KeyF2},
		{"13", KeyF3},
		{"14", KeyF4},
	}
	for _, tt := range tests {
		seq := append([]byte{0x1b, '['}, []byte(tt.params)...)
		seq = append(seq, '~')
		k := requireOneKey(t, feed(seq...))
		if k.Code != tt.code {
			t.Fatalf("params %s: got %v, want %v", tt.params, k.Code, tt.code)
		}
	}
}
