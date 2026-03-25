package text

// RuneWidth returns the display width of a rune in terminal columns.
func RuneWidth(r rune) int {
	// Control characters.
	if r < 0x20 || r == 0x7F {
		return 0
	}
	// ASCII printable.
	if r <= 0x7E {
		return 1
	}
	// CJK and other wide character ranges.
	if isWide(r) {
		return 2
	}
	return 1
}

func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F: // Hangul Jamo
		return true
	case r >= 0x2E80 && r <= 0x303E: // CJK Radicals, Kangxi, CJK Symbols
		return true
	case r >= 0x3040 && r <= 0x33BF: // Hiragana, Katakana, Bopomofo, Hangul Compat, Kanbun, CJK Compat
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Unified Extension A
		return true
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified
		return true
	case r >= 0xA000 && r <= 0xA4CF: // Yi
		return true
	case r >= 0xAC00 && r <= 0xD7AF: // Hangul Syllables
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
		return true
	case r >= 0xFE30 && r <= 0xFE6F: // CJK Compatibility Forms
		return true
	case r >= 0xFF01 && r <= 0xFF60: // Fullwidth Latin
		return true
	case r >= 0xFFE0 && r <= 0xFFE6: // Fullwidth Signs
		return true
	case r >= 0x20000 && r <= 0x2FFFD: // CJK Extension B+
		return true
	case r >= 0x30000 && r <= 0x3FFFD: // CJK Extension G+
		return true
	}
	return false
}

// StringWidth returns the total display width of a string.
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}

// StyledLineWidth returns the total display width of a styled line.
func StyledLineWidth(line StyledLine) int {
	w := 0
	for _, sp := range line {
		w += StringWidth(sp.Text)
	}
	return w
}

// Truncate truncates a string to fit within maxWidth display columns.
// If a wide character would cause the string to exceed maxWidth, it is omitted.
func Truncate(s string, maxWidth int) string {
	w := 0
	for i, r := range s {
		rw := RuneWidth(r)
		if w+rw > maxWidth {
			return s[:i]
		}
		w += rw
	}
	return s
}

// StyledTruncate truncates a styled line to fit within maxWidth display columns.
func StyledTruncate(line StyledLine, maxWidth int) StyledLine {
	if maxWidth <= 0 {
		return nil
	}
	var result StyledLine
	remaining := maxWidth
	for _, sp := range line {
		if remaining <= 0 {
			break
		}
		truncated := Truncate(sp.Text, remaining)
		if len(truncated) > 0 {
			result = append(result, StyledSpan{Text: truncated, Style: sp.Style})
		}
		remaining -= StringWidth(truncated)
	}
	return result
}
