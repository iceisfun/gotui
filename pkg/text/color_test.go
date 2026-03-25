package text

import "testing"

func TestDefault(t *testing.T) {
	c := Default()
	if !c.IsDefault() {
		t.Fatal("Default() should be the default color")
	}
}

func TestColor16(t *testing.T) {
	c := Color16(5)
	if c.IsDefault() {
		t.Fatal("Color16 should not be default")
	}
	if c.FgSGR() != "35" {
		t.Fatalf("FgSGR() = %q, want %q", c.FgSGR(), "35")
	}
	if c.BgSGR() != "45" {
		t.Fatalf("BgSGR() = %q, want %q", c.BgSGR(), "45")
	}
}

func TestColor16Bright(t *testing.T) {
	c := Color16(10) // bright green
	if c.FgSGR() != "92" {
		t.Fatalf("FgSGR() = %q, want %q", c.FgSGR(), "92")
	}
	if c.BgSGR() != "102" {
		t.Fatalf("BgSGR() = %q, want %q", c.BgSGR(), "102")
	}
}

func TestColor256(t *testing.T) {
	c := Color256(200)
	if c.FgSGR() != "38;5;200" {
		t.Fatalf("FgSGR() = %q, want %q", c.FgSGR(), "38;5;200")
	}
	if c.BgSGR() != "48;5;200" {
		t.Fatalf("BgSGR() = %q, want %q", c.BgSGR(), "48;5;200")
	}
}

func TestRGB(t *testing.T) {
	c := RGB(0xff, 0x88, 0x00)
	if c.FgSGR() != "38;2;255;136;0" {
		t.Fatalf("FgSGR() = %q, want %q", c.FgSGR(), "38;2;255;136;0")
	}
	if c.BgSGR() != "48;2;255;136;0" {
		t.Fatalf("BgSGR() = %q, want %q", c.BgSGR(), "48;2;255;136;0")
	}
}

func TestDefaultSGR(t *testing.T) {
	c := Default()
	if c.FgSGR() != "39" {
		t.Fatalf("FgSGR() = %q, want %q", c.FgSGR(), "39")
	}
	if c.BgSGR() != "49" {
		t.Fatalf("BgSGR() = %q, want %q", c.BgSGR(), "49")
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b Color
		want bool
	}{
		{"same default", Default(), Default(), true},
		{"same color16", Color16(3), Color16(3), true},
		{"diff color16", Color16(3), Color16(4), false},
		{"same color256", Color256(100), Color256(100), true},
		{"diff color256", Color256(100), Color256(101), false},
		{"same rgb", RGB(1, 2, 3), RGB(1, 2, 3), true},
		{"diff rgb", RGB(1, 2, 3), RGB(1, 2, 4), false},
		{"default vs color16", Default(), Color16(0), false},
		{"color16 vs color256", Color16(5), Color256(5), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNamedColors(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		fgSGR string
	}{
		{"Black", Black(), "30"},
		{"Red", Red(), "31"},
		{"Green", Green(), "32"},
		{"Yellow", Yellow(), "33"},
		{"Blue", Blue(), "34"},
		{"Magenta", Magenta(), "35"},
		{"Cyan", Cyan(), "36"},
		{"White", White(), "37"},
		{"BrightBlack", BrightBlack(), "90"},
		{"BrightRed", BrightRed(), "91"},
		{"BrightGreen", BrightGreen(), "92"},
		{"BrightYellow", BrightYellow(), "93"},
		{"BrightBlue", BrightBlue(), "94"},
		{"BrightMagenta", BrightMagenta(), "95"},
		{"BrightCyan", BrightCyan(), "96"},
		{"BrightWhite", BrightWhite(), "97"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.color.FgSGR(); got != tt.fgSGR {
				t.Fatalf("FgSGR() = %q, want %q", got, tt.fgSGR)
			}
		})
	}
}

func TestFgSGRTable(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{"color16 index 0", Color16(0), "30"},
		{"color16 index 7", Color16(7), "37"},
		{"color16 index 8", Color16(8), "90"},
		{"color16 index 15", Color16(15), "97"},
		{"color256 index 0", Color256(0), "38;5;0"},
		{"color256 index 255", Color256(255), "38;5;255"},
		{"rgb black", RGB(0, 0, 0), "38;2;0;0;0"},
		{"rgb white", RGB(255, 255, 255), "38;2;255;255;255"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.color.FgSGR(); got != tt.want {
				t.Fatalf("FgSGR() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBgSGRTable(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{"color16 index 0", Color16(0), "40"},
		{"color16 index 7", Color16(7), "47"},
		{"color16 index 8", Color16(8), "100"},
		{"color16 index 15", Color16(15), "107"},
		{"color256 index 0", Color256(0), "48;5;0"},
		{"rgb", RGB(10, 20, 30), "48;2;10;20;30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.color.BgSGR(); got != tt.want {
				t.Fatalf("BgSGR() = %q, want %q", got, tt.want)
			}
		})
	}
}
