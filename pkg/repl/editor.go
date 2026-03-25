package repl

import (
	"strings"
	"unicode"
)

// Editor is a multiline text editor with cursor navigation.
// It stores text as a slice of lines (simple and predictable for a REPL).
type Editor struct {
	lines  [][]rune // One entry per line.
	curRow int      // Cursor row (0-indexed).
	curCol int      // Cursor column (0-indexed, clamped to line length).
}

// NewEditor creates an empty editor.
func NewEditor() *Editor {
	return &Editor{
		lines: [][]rune{{}},
	}
}

// Text returns the full editor content as a string.
func (e *Editor) Text() string {
	var b strings.Builder
	for i, line := range e.lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(string(line))
	}
	return b.String()
}

// SetText replaces the editor content.
func (e *Editor) SetText(s string) {
	raw := strings.Split(s, "\n")
	e.lines = make([][]rune, len(raw))
	for i, r := range raw {
		e.lines[i] = []rune(r)
	}
	e.curRow = len(e.lines) - 1
	e.curCol = len(e.lines[e.curRow])
}

// Clear resets the editor to empty.
func (e *Editor) Clear() {
	e.lines = [][]rune{{}}
	e.curRow = 0
	e.curCol = 0
}

// Lines returns the number of lines.
func (e *Editor) Lines() int { return len(e.lines) }

// Line returns the content of line i as a string.
func (e *Editor) Line(i int) string {
	if i < 0 || i >= len(e.lines) {
		return ""
	}
	return string(e.lines[i])
}

// CursorPos returns (row, col).
func (e *Editor) CursorPos() (row, col int) {
	return e.curRow, e.curCol
}

// InsertRune inserts a rune at the cursor position.
func (e *Editor) InsertRune(r rune) {
	line := e.lines[e.curRow]
	col := e.clampCol()

	newLine := make([]rune, len(line)+1)
	copy(newLine, line[:col])
	newLine[col] = r
	copy(newLine[col+1:], line[col:])
	e.lines[e.curRow] = newLine
	e.curCol = col + 1
}

// InsertNewline splits the current line at the cursor.
func (e *Editor) InsertNewline() {
	col := e.clampCol()
	line := e.lines[e.curRow]

	// Split into before and after cursor.
	before := make([]rune, col)
	copy(before, line[:col])
	after := make([]rune, len(line)-col)
	copy(after, line[col:])

	// Insert new line.
	e.lines[e.curRow] = before
	newLines := make([][]rune, len(e.lines)+1)
	copy(newLines, e.lines[:e.curRow+1])
	newLines[e.curRow+1] = after
	copy(newLines[e.curRow+2:], e.lines[e.curRow+1:])
	e.lines = newLines

	e.curRow++
	e.curCol = 0
}

// InsertString inserts a string (possibly multiline) at the cursor.
func (e *Editor) InsertString(s string) {
	for _, r := range s {
		if r == '\n' {
			e.InsertNewline()
		} else {
			e.InsertRune(r)
		}
	}
}

// Backspace deletes the character before the cursor.
func (e *Editor) Backspace() {
	col := e.clampCol()

	if col > 0 {
		line := e.lines[e.curRow]
		e.lines[e.curRow] = append(line[:col-1], line[col:]...)
		e.curCol = col - 1
	} else if e.curRow > 0 {
		// Join with previous line.
		prevLen := len(e.lines[e.curRow-1])
		e.lines[e.curRow-1] = append(e.lines[e.curRow-1], e.lines[e.curRow]...)
		e.lines = append(e.lines[:e.curRow], e.lines[e.curRow+1:]...)
		e.curRow--
		e.curCol = prevLen
	}
}

// Delete deletes the character at the cursor.
func (e *Editor) Delete() {
	col := e.clampCol()
	line := e.lines[e.curRow]

	if col < len(line) {
		e.lines[e.curRow] = append(line[:col], line[col+1:]...)
	} else if e.curRow < len(e.lines)-1 {
		// Join with next line.
		e.lines[e.curRow] = append(line, e.lines[e.curRow+1]...)
		e.lines = append(e.lines[:e.curRow+1], e.lines[e.curRow+2:]...)
	}
}

// MoveLeft moves the cursor one character left.
func (e *Editor) MoveLeft() {
	col := e.clampCol()
	if col > 0 {
		e.curCol = col - 1
	} else if e.curRow > 0 {
		e.curRow--
		e.curCol = len(e.lines[e.curRow])
	}
}

// MoveRight moves the cursor one character right.
func (e *Editor) MoveRight() {
	col := e.clampCol()
	if col < len(e.lines[e.curRow]) {
		e.curCol = col + 1
	} else if e.curRow < len(e.lines)-1 {
		e.curRow++
		e.curCol = 0
	}
}

// MoveUp moves the cursor one line up.
func (e *Editor) MoveUp() {
	if e.curRow > 0 {
		e.curRow--
		e.curCol = min(e.curCol, len(e.lines[e.curRow]))
	}
}

// MoveDown moves the cursor one line down.
func (e *Editor) MoveDown() {
	if e.curRow < len(e.lines)-1 {
		e.curRow++
		e.curCol = min(e.curCol, len(e.lines[e.curRow]))
	}
}

// MoveHome moves the cursor to the beginning of the line.
func (e *Editor) MoveHome() {
	e.curCol = 0
}

// MoveEnd moves the cursor to the end of the line.
func (e *Editor) MoveEnd() {
	e.curCol = len(e.lines[e.curRow])
}

// MoveWordLeft moves the cursor one word left.
func (e *Editor) MoveWordLeft() {
	col := e.clampCol()
	line := e.lines[e.curRow]

	if col == 0 {
		e.MoveLeft()
		return
	}

	// Skip whitespace, then skip word characters.
	i := col - 1
	for i > 0 && unicode.IsSpace(line[i]) {
		i--
	}
	for i > 0 && !unicode.IsSpace(line[i-1]) {
		i--
	}
	e.curCol = i
}

// MoveWordRight moves the cursor one word right.
func (e *Editor) MoveWordRight() {
	col := e.clampCol()
	line := e.lines[e.curRow]

	if col >= len(line) {
		e.MoveRight()
		return
	}

	// Skip word characters, then skip whitespace.
	i := col
	for i < len(line) && !unicode.IsSpace(line[i]) {
		i++
	}
	for i < len(line) && unicode.IsSpace(line[i]) {
		i++
	}
	e.curCol = i
}

// DeleteWordBack deletes from cursor back to the start of the previous word.
func (e *Editor) DeleteWordBack() {
	col := e.clampCol()
	if col == 0 {
		e.Backspace()
		return
	}

	line := e.lines[e.curRow]
	i := col - 1
	for i > 0 && unicode.IsSpace(line[i]) {
		i--
	}
	for i > 0 && !unicode.IsSpace(line[i-1]) {
		i--
	}

	e.lines[e.curRow] = append(line[:i], line[col:]...)
	e.curCol = i
}

// DeleteToEnd deletes from cursor to end of line.
func (e *Editor) DeleteToEnd() {
	col := e.clampCol()
	e.lines[e.curRow] = e.lines[e.curRow][:col]
}

// DeleteToStart deletes from start of line to cursor.
func (e *Editor) DeleteToStart() {
	col := e.clampCol()
	e.lines[e.curRow] = e.lines[e.curRow][col:]
	e.curCol = 0
}

func (e *Editor) clampCol() int {
	lineLen := len(e.lines[e.curRow])
	if e.curCol > lineLen {
		e.curCol = lineLen
	}
	return e.curCol
}
