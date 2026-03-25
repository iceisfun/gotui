package text

// Attr represents text attributes as a bitmask.
type Attr uint8

const (
	AttrBold Attr = 1 << iota
	AttrDim
	AttrItalic
	AttrUnderline
	AttrBlink
	AttrReverse
	AttrStrikethrough
)

// Style combines foreground color, background color, and text attributes.
type Style struct {
	Fg    Color
	Bg    Color
	Attrs Attr
}

func (s Style) WithFg(c Color) Style    { s.Fg = c; return s }
func (s Style) WithBg(c Color) Style    { s.Bg = c; return s }
func (s Style) WithAttr(a Attr) Style   { s.Attrs |= a; return s }
func (s Style) WithoutAttr(a Attr) Style { s.Attrs &^= a; return s }

func (s Style) Bold() Style          { return s.WithAttr(AttrBold) }
func (s Style) Dim() Style           { return s.WithAttr(AttrDim) }
func (s Style) Italic() Style        { return s.WithAttr(AttrItalic) }
func (s Style) Underline() Style     { return s.WithAttr(AttrUnderline) }
func (s Style) Reverse() Style       { return s.WithAttr(AttrReverse) }
func (s Style) Strikethrough() Style { return s.WithAttr(AttrStrikethrough) }

// Equal reports whether two styles are identical.
func (s Style) Equal(other Style) bool {
	return s.Fg.Equal(other.Fg) && s.Bg.Equal(other.Bg) && s.Attrs == other.Attrs
}
