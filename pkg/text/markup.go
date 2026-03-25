package text

import "strings"

// Parse parses a markup string into a StyledLine.
//
// Syntax:
//
//	[bold red on blue]styled text[/]
//	[italic #ff8800]rgb foreground[/]
//	[dim on #003366]rgb background[/]
//	plain text with no tags
//
// Inside [...], space-separated tokens are parsed:
//   - Color names set foreground: red, green, blue, yellow, cyan, magenta, white, black,
//     bright_red, bright_green, etc.
//   - "on <color>" sets background.
//   - Attribute names: bold, dim, italic, underline, blink, reverse, strikethrough.
//   - Hex colors #RRGGBB set foreground, or background after "on".
//   - [/] resets to default style.
//
// Tags nest: [bold][red]text[/][/] — each [/] pops one style from the stack.
func Parse(markup string) StyledLine {
	var line StyledLine
	var styleStack []Style
	current := Style{}

	i := 0
	for i < len(markup) {
		// Look for tag open.
		idx := strings.IndexByte(markup[i:], '[')
		if idx < 0 {
			// Rest is plain text.
			if i < len(markup) {
				line = append(line, StyledSpan{Text: markup[i:], Style: current})
			}
			break
		}

		// Emit text before the tag.
		if idx > 0 {
			line = append(line, StyledSpan{Text: markup[i : i+idx], Style: current})
		}
		i += idx + 1 // skip '['

		// Find closing ']'.
		end := strings.IndexByte(markup[i:], ']')
		if end < 0 {
			// No closing bracket — treat rest as literal text.
			line = append(line, StyledSpan{Text: markup[i-1:], Style: current})
			break
		}

		tagContent := markup[i : i+end]
		i += end + 1 // skip ']'

		// [/] pops style stack.
		if tagContent == "/" {
			if len(styleStack) > 0 {
				current = styleStack[len(styleStack)-1]
				styleStack = styleStack[:len(styleStack)-1]
			} else {
				current = Style{}
			}
			continue
		}

		// Push current style and parse new tag.
		styleStack = append(styleStack, current)
		current = parseTag(tagContent, current)
	}

	return line
}

func parseTag(content string, base Style) Style {
	s := base
	tokens := strings.Fields(content)
	expectBg := false

	for _, tok := range tokens {
		if tok == "on" {
			expectBg = true
			continue
		}

		// Check for hex color.
		if len(tok) == 7 && tok[0] == '#' {
			c := parseHexColor(tok)
			if expectBg {
				s.Bg = c
				expectBg = false
			} else {
				s.Fg = c
			}
			continue
		}

		// Check for attribute.
		if attr, ok := attrByName(tok); ok {
			s.Attrs |= attr
			continue
		}

		// Check for color name.
		if c, ok := colorByName(tok); ok {
			if expectBg {
				s.Bg = c
				expectBg = false
			} else {
				s.Fg = c
			}
			continue
		}
	}
	return s
}

func parseHexColor(hex string) Color {
	// #RRGGBB
	r := hexByte(hex[1], hex[2])
	g := hexByte(hex[3], hex[4])
	b := hexByte(hex[5], hex[6])
	return RGB(r, g, b)
}

func hexByte(hi, lo byte) uint8 {
	return hexDigit(hi)<<4 | hexDigit(lo)
}

func hexDigit(b byte) uint8 {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	}
	return 0
}

func attrByName(name string) (Attr, bool) {
	switch strings.ToLower(name) {
	case "bold":
		return AttrBold, true
	case "dim":
		return AttrDim, true
	case "italic":
		return AttrItalic, true
	case "underline":
		return AttrUnderline, true
	case "blink":
		return AttrBlink, true
	case "reverse":
		return AttrReverse, true
	case "strikethrough":
		return AttrStrikethrough, true
	}
	return 0, false
}

func colorByName(name string) (Color, bool) {
	switch strings.ToLower(name) {
	case "black":
		return Black(), true
	case "red":
		return Red(), true
	case "green":
		return Green(), true
	case "yellow":
		return Yellow(), true
	case "blue":
		return Blue(), true
	case "magenta":
		return Magenta(), true
	case "cyan":
		return Cyan(), true
	case "white":
		return White(), true
	case "bright_black", "brightblack":
		return BrightBlack(), true
	case "bright_red", "brightred":
		return BrightRed(), true
	case "bright_green", "brightgreen":
		return BrightGreen(), true
	case "bright_yellow", "brightyellow":
		return BrightYellow(), true
	case "bright_blue", "brightblue":
		return BrightBlue(), true
	case "bright_magenta", "brightmagenta":
		return BrightMagenta(), true
	case "bright_cyan", "brightcyan":
		return BrightCyan(), true
	case "bright_white", "brightwhite":
		return BrightWhite(), true
	case "default":
		return Default(), true
	}
	return Color{}, false
}
