package text

// StyledSpan is a contiguous run of text with a single style.
type StyledSpan struct {
	Text  string
	Style Style
}

// StyledLine is a sequence of styled spans forming one visual line.
type StyledLine []StyledSpan

// Len returns the total character width of the line.
func (l StyledLine) Len() int {
	n := 0
	for _, sp := range l {
		n += len([]rune(sp.Text))
	}
	return n
}

// PlainText returns the unstyled text content.
func (l StyledLine) PlainText() string {
	if len(l) == 1 {
		return l[0].Text
	}
	n := 0
	for _, sp := range l {
		n += len(sp.Text)
	}
	buf := make([]byte, 0, n)
	for _, sp := range l {
		buf = append(buf, sp.Text...)
	}
	return string(buf)
}

// Append adds a styled span to the line.
func (l StyledLine) Append(text string, s Style) StyledLine {
	return append(l, StyledSpan{Text: text, Style: s})
}

// Plain creates a StyledLine from unstyled text.
func Plain(s string) StyledLine {
	return StyledLine{{Text: s}}
}

// Styled creates a single-span StyledLine.
func Styled(s string, style Style) StyledLine {
	return StyledLine{{Text: s, Style: style}}
}
