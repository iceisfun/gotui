package text

import "strings"

// Wrap wraps a plain string at word boundaries to fit within width columns.
// If a single word exceeds width, it is broken mid-word. Existing newlines
// are respected. Width measurement is rune-width-aware.
func Wrap(s string, width int) []string {
	if width <= 0 {
		return nil
	}
	if s == "" {
		return []string{""}
	}

	var result []string
	for _, paragraph := range strings.Split(s, "\n") {
		result = append(result, wrapParagraph(paragraph, width)...)
	}
	return result
}

func wrapParagraph(s string, width int) []string {
	if s == "" {
		return []string{""}
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var line strings.Builder
	lineWidth := 0

	for i, word := range words {
		wordWidth := StringWidth(word)

		// If the word itself exceeds width, break it.
		if wordWidth > width {
			// Flush current line if non-empty.
			if lineWidth > 0 {
				lines = append(lines, line.String())
				line.Reset()
				lineWidth = 0
			}
			lines = append(lines, breakWord(word, width)...)
			// The last element of breakWord is the remainder; use it as the current line.
			last := lines[len(lines)-1]
			lines = lines[:len(lines)-1]
			line.WriteString(last)
			lineWidth = StringWidth(last)
			continue
		}

		if i == 0 {
			line.WriteString(word)
			lineWidth = wordWidth
			continue
		}

		// Check if word fits on current line with a space.
		if lineWidth+1+wordWidth <= width {
			line.WriteByte(' ')
			line.WriteString(word)
			lineWidth += 1 + wordWidth
		} else {
			lines = append(lines, line.String())
			line.Reset()
			line.WriteString(word)
			lineWidth = wordWidth
		}
	}

	if line.Len() > 0 || lineWidth == 0 {
		lines = append(lines, line.String())
	}
	return lines
}

// breakWord breaks a word that exceeds width into multiple segments.
func breakWord(word string, width int) []string {
	var parts []string
	var cur strings.Builder
	curWidth := 0

	for _, r := range word {
		rw := RuneWidth(r)
		if curWidth+rw > width && cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
			curWidth = 0
		}
		cur.WriteRune(r)
		curWidth += rw
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// WrapStyled wraps a styled line at word boundaries to fit within width columns,
// preserving style information across line breaks.
func WrapStyled(line StyledLine, width int) []StyledLine {
	if width <= 0 {
		return nil
	}
	if len(line) == 0 {
		return []StyledLine{{}}
	}

	// We iterate through spans rune-by-rune, tracking current output line width.
	var result []StyledLine
	var curLine StyledLine
	var curSpanText strings.Builder
	var curStyle Style
	curLineWidth := 0
	haveSpanStyle := false

	// wordBuf accumulates the current word (runes + their style spans).
	type styledRune struct {
		r     rune
		style Style
		width int
	}
	var wordBuf []styledRune
	wordWidth := 0

	flushSpan := func() {
		if curSpanText.Len() > 0 {
			curLine = append(curLine, StyledSpan{Text: curSpanText.String(), Style: curStyle})
			curSpanText.Reset()
		}
		haveSpanStyle = false
	}

	appendRune := func(r rune, style Style, rw int) {
		if haveSpanStyle && style != curStyle {
			flushSpan()
		}
		if !haveSpanStyle {
			curStyle = style
			haveSpanStyle = true
		}
		curSpanText.WriteRune(r)
		curLineWidth += rw
	}

	newLine := func() {
		flushSpan()
		result = append(result, curLine)
		curLine = nil
		curLineWidth = 0
	}

	flushWord := func() {
		if len(wordBuf) == 0 {
			return
		}

		// If the word fits on the current line (with space if needed).
		spaceNeeded := 0
		if curLineWidth > 0 {
			spaceNeeded = 1
		}

		if wordWidth <= width && curLineWidth+spaceNeeded+wordWidth <= width {
			// Fits: add space if needed, then the word.
			if spaceNeeded > 0 {
				appendRune(' ', wordBuf[0].style, 1)
			}
			for _, sr := range wordBuf {
				appendRune(sr.r, sr.style, sr.width)
			}
		} else if wordWidth <= width {
			// Doesn't fit on current line but fits on a new line.
			if curLineWidth > 0 {
				newLine()
			}
			for _, sr := range wordBuf {
				appendRune(sr.r, sr.style, sr.width)
			}
		} else {
			// Word exceeds width; break it.
			if curLineWidth > 0 {
				newLine()
			}
			for _, sr := range wordBuf {
				if curLineWidth+sr.width > width && curLineWidth > 0 {
					newLine()
				}
				appendRune(sr.r, sr.style, sr.width)
			}
		}
		wordBuf = wordBuf[:0]
		wordWidth = 0
	}

	for _, sp := range line {
		for _, r := range sp.Text {
			if r == '\n' {
				flushWord()
				newLine()
				continue
			}
			rw := RuneWidth(r)
			if r == ' ' || r == '\t' {
				flushWord()
				continue
			}
			wordBuf = append(wordBuf, styledRune{r: r, style: sp.Style, width: rw})
			wordWidth += rw
		}
	}
	flushWord()
	flushSpan()
	result = append(result, curLine)

	return result
}
