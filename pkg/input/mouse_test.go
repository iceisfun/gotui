package input

import "testing"

func TestParseSGRMouseLeftPress(t *testing.T) {
	me, ok := parseSGRMouse("0;10;20", 'M')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseLeft {
		t.Fatalf("button = %v, want MouseLeft", me.Button)
	}
	// Coordinates are 1-based, so 10 -> 9, 20 -> 19.
	if me.X != 9 || me.Y != 19 {
		t.Fatalf("coords = (%d,%d), want (9,19)", me.X, me.Y)
	}
}

func TestParseSGRMouseMiddlePress(t *testing.T) {
	me, ok := parseSGRMouse("1;5;5", 'M')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseMiddle {
		t.Fatalf("button = %v, want MouseMiddle", me.Button)
	}
}

func TestParseSGRMouseRightPress(t *testing.T) {
	me, ok := parseSGRMouse("2;5;5", 'M')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseRight {
		t.Fatalf("button = %v, want MouseRight", me.Button)
	}
}

func TestParseSGRMouseRelease(t *testing.T) {
	me, ok := parseSGRMouse("0;10;20", 'm')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseRelease {
		t.Fatalf("button = %v, want MouseRelease", me.Button)
	}
}

func TestParseSGRMouseWheelUp(t *testing.T) {
	me, ok := parseSGRMouse("64;10;20", 'M')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseWheelUp {
		t.Fatalf("button = %v, want MouseWheelUp", me.Button)
	}
}

func TestParseSGRMouseWheelDown(t *testing.T) {
	me, ok := parseSGRMouse("65;10;20", 'M')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseWheelDown {
		t.Fatalf("button = %v, want MouseWheelDown", me.Button)
	}
}

func TestParseSGRMouseModifiers(t *testing.T) {
	tests := []struct {
		name   string
		params string
		mod    ModMask
	}{
		// Shift = bit 2 (value 4)
		{"Shift", "4;1;1", ModShift},
		// Alt = bit 3 (value 8)
		{"Alt", "8;1;1", ModAlt},
		// Ctrl = bit 4 (value 16)
		{"Ctrl", "16;1;1", ModCtrl},
		// Shift+Alt = 4+8 = 12
		{"Shift+Alt", "12;1;1", ModShift | ModAlt},
		// Shift+Ctrl = 4+16 = 20
		{"Shift+Ctrl", "20;1;1", ModShift | ModCtrl},
		// All = 4+8+16 = 28
		{"All", "28;1;1", ModShift | ModAlt | ModCtrl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			me, ok := parseSGRMouse(tt.params, 'M')
			if !ok {
				t.Fatal("expected ok")
			}
			if me.Mod != tt.mod {
				t.Fatalf("mod = %v, want %v", me.Mod, tt.mod)
			}
		})
	}
}

func TestParseSGRMouseMotion(t *testing.T) {
	// Motion = bit 5 (value 32) + left button (0)
	me, ok := parseSGRMouse("32;5;5", 'M')
	if !ok {
		t.Fatal("expected ok")
	}
	if me.Button != MouseMotion {
		t.Fatalf("button = %v, want MouseMotion", me.Button)
	}
}

func TestParseSGRMouseInvalid(t *testing.T) {
	tests := []struct {
		name   string
		params string
	}{
		{"too few parts", "0;10"},
		{"too many parts", "0;10;20;30"},
		{"non-numeric", "a;10;20"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := parseSGRMouse(tt.params, 'M')
			if ok {
				t.Fatal("expected !ok for invalid input")
			}
		})
	}
}
