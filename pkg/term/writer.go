package term

import (
	"bytes"
	"fmt"
	"os"

	"github.com/iceisfun/gotui/pkg/text"
)

// Writer is a buffered writer that builds ANSI escape sequences
// and flushes them to the terminal in a single write.
type Writer struct {
	buf bytes.Buffer
	out *os.File
}

// NewWriter creates a new Writer targeting the given file.
func NewWriter(out *os.File) *Writer {
	return &Writer{out: out}
}

// WriteRaw writes a raw string to the buffer without interpretation.
func (w *Writer) WriteRaw(s string) {
	w.buf.WriteString(s)
}

// WriteString writes text to the buffer.
func (w *Writer) WriteString(s string) {
	w.buf.WriteString(s)
}

// WriteRune writes a single rune to the buffer.
func (w *Writer) WriteRune(r rune) {
	w.buf.WriteRune(r)
}

// MoveTo moves the cursor to row, col (1-based).
func (w *Writer) MoveTo(row, col int) {
	fmt.Fprintf(&w.buf, "\x1b[%d;%dH", row, col)
}

// MoveUp moves the cursor up by n rows.
func (w *Writer) MoveUp(n int) {
	if n > 0 {
		fmt.Fprintf(&w.buf, "\x1b[%dA", n)
	}
}

// MoveDown moves the cursor down by n rows.
func (w *Writer) MoveDown(n int) {
	if n > 0 {
		fmt.Fprintf(&w.buf, "\x1b[%dB", n)
	}
}

// MoveRight moves the cursor right by n columns.
func (w *Writer) MoveRight(n int) {
	if n > 0 {
		fmt.Fprintf(&w.buf, "\x1b[%dC", n)
	}
}

// MoveLeft moves the cursor left by n columns.
func (w *Writer) MoveLeft(n int) {
	if n > 0 {
		fmt.Fprintf(&w.buf, "\x1b[%dD", n)
	}
}

// ClearScreen clears the entire screen.
func (w *Writer) ClearScreen() {
	w.buf.WriteString("\x1b[2J")
}

// ClearLine clears the current line.
func (w *Writer) ClearLine() {
	w.buf.WriteString("\x1b[2K")
}

// ClearToEndOfLine clears from cursor to end of line.
func (w *Writer) ClearToEndOfLine() {
	w.buf.WriteString("\x1b[0K")
}

// HideCursor hides the cursor.
func (w *Writer) HideCursor() {
	w.buf.WriteString("\x1b[?25l")
}

// ShowCursor shows the cursor.
func (w *Writer) ShowCursor() {
	w.buf.WriteString("\x1b[?25h")
}

// ResetStyle resets all text attributes and colors.
func (w *Writer) ResetStyle() {
	w.buf.WriteString("\x1b[0m")
}

// SetStyle applies a complete style (fg + bg + attrs) as SGR sequences.
func (w *Writer) SetStyle(s text.Style) {
	w.ResetStyle()
	if !s.Fg.IsDefault() {
		fmt.Fprintf(&w.buf, "\x1b[%sm", s.Fg.FgSGR())
	}
	if !s.Bg.IsDefault() {
		fmt.Fprintf(&w.buf, "\x1b[%sm", s.Bg.BgSGR())
	}
	if s.Attrs&text.AttrBold != 0 {
		w.buf.WriteString("\x1b[1m")
	}
	if s.Attrs&text.AttrDim != 0 {
		w.buf.WriteString("\x1b[2m")
	}
	if s.Attrs&text.AttrItalic != 0 {
		w.buf.WriteString("\x1b[3m")
	}
	if s.Attrs&text.AttrUnderline != 0 {
		w.buf.WriteString("\x1b[4m")
	}
	if s.Attrs&text.AttrBlink != 0 {
		w.buf.WriteString("\x1b[5m")
	}
	if s.Attrs&text.AttrReverse != 0 {
		w.buf.WriteString("\x1b[7m")
	}
	if s.Attrs&text.AttrStrikethrough != 0 {
		w.buf.WriteString("\x1b[9m")
	}
}

// SetFg sets only the foreground color.
func (w *Writer) SetFg(c text.Color) {
	fmt.Fprintf(&w.buf, "\x1b[%sm", c.FgSGR())
}

// SetBg sets only the background color.
func (w *Writer) SetBg(c text.Color) {
	fmt.Fprintf(&w.buf, "\x1b[%sm", c.BgSGR())
}

// Flush writes the accumulated buffer to the terminal and resets it.
func (w *Writer) Flush() error {
	if w.buf.Len() == 0 {
		return nil
	}
	_, err := w.out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}
