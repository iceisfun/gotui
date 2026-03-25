package text

import "fmt"

type colorType uint8

const (
	colorDefault colorType = iota
	color16
	color256
	colorRGB
)

// Color represents a terminal color. Zero value is the default color.
type Color struct {
	typ   colorType
	value uint32
}

// Default returns the terminal's default color.
func Default() Color { return Color{} }

// Color16 returns a standard 16-color palette color (0–15).
func Color16(n uint8) Color { return Color{typ: color16, value: uint32(n)} }

// Color256 returns an extended 256-color palette color (0–255).
func Color256(n uint8) Color { return Color{typ: color256, value: uint32(n)} }

// RGB returns a 24-bit truecolor.
func RGB(r, g, b uint8) Color {
	return Color{typ: colorRGB, value: uint32(r)<<16 | uint32(g)<<8 | uint32(b)}
}

// Standard 16 colors.
func Black() Color         { return Color16(0) }
func Red() Color           { return Color16(1) }
func Green() Color         { return Color16(2) }
func Yellow() Color        { return Color16(3) }
func Blue() Color          { return Color16(4) }
func Magenta() Color       { return Color16(5) }
func Cyan() Color          { return Color16(6) }
func White() Color         { return Color16(7) }
func BrightBlack() Color   { return Color16(8) }
func BrightRed() Color     { return Color16(9) }
func BrightGreen() Color   { return Color16(10) }
func BrightYellow() Color  { return Color16(11) }
func BrightBlue() Color    { return Color16(12) }
func BrightMagenta() Color { return Color16(13) }
func BrightCyan() Color    { return Color16(14) }
func BrightWhite() Color   { return Color16(15) }

// IsDefault reports whether c is the terminal's default color.
func (c Color) IsDefault() bool { return c.typ == colorDefault }

// FgSGR returns the SGR parameter string for using this color as foreground.
func (c Color) FgSGR() string {
	switch c.typ {
	case colorDefault:
		return "39"
	case color16:
		n := c.value
		if n < 8 {
			return fmt.Sprintf("%d", 30+n)
		}
		return fmt.Sprintf("%d", 90+(n-8))
	case color256:
		return fmt.Sprintf("38;5;%d", c.value)
	case colorRGB:
		r := (c.value >> 16) & 0xff
		g := (c.value >> 8) & 0xff
		b := c.value & 0xff
		return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	}
	return "39"
}

// BgSGR returns the SGR parameter string for using this color as background.
func (c Color) BgSGR() string {
	switch c.typ {
	case colorDefault:
		return "49"
	case color16:
		n := c.value
		if n < 8 {
			return fmt.Sprintf("%d", 40+n)
		}
		return fmt.Sprintf("%d", 100+(n-8))
	case color256:
		return fmt.Sprintf("48;5;%d", c.value)
	case colorRGB:
		r := (c.value >> 16) & 0xff
		g := (c.value >> 8) & 0xff
		b := c.value & 0xff
		return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	}
	return "49"
}

// Equal reports whether two colors are identical.
func (c Color) Equal(other Color) bool {
	return c.typ == other.typ && c.value == other.value
}
