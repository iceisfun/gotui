package term

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Terminal manages raw mode and provides I/O handles for the controlling terminal.
type Terminal struct {
	in       *os.File
	out      *os.File
	fd       int
	oldState *term.State
	reader   *Reader
	writer   *Writer
}

// Open puts the terminal into raw mode and returns a Terminal.
// The caller must call Close to restore the terminal state.
func Open() (*Terminal, error) {
	in := os.Stdin
	out := os.Stdout
	fd := int(in.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("term: make raw: %w", err)
	}

	t := &Terminal{
		in:       in,
		out:      out,
		fd:       fd,
		oldState: oldState,
		writer:   NewWriter(out),
		reader:   NewReader(in),
	}

	// Enable bracketed paste mode.
	t.writer.WriteRaw("\x1b[?2004h")
	// Enable SGR mouse tracking (all events + SGR encoding).
	// 1003 = all-motion tracking (press/release/drag/motion).
	// 1006 = SGR encoding (supports coordinates > 223).
	t.writer.WriteRaw("\x1b[?1003h")
	t.writer.WriteRaw("\x1b[?1006h")
	// Switch to alternate screen buffer.
	t.writer.WriteRaw("\x1b[?1049h")
	// Hide cursor initially.
	t.writer.HideCursor()
	if err := t.writer.Flush(); err != nil {
		_ = term.Restore(fd, oldState)
		return nil, fmt.Errorf("term: init sequences: %w", err)
	}

	return t, nil
}

// Close restores the terminal to its original state.
func (t *Terminal) Close() error {
	t.reader.Stop()

	// Disable SGR mouse.
	t.writer.WriteRaw("\x1b[?1006l")
	t.writer.WriteRaw("\x1b[?1003l")
	// Disable bracketed paste.
	t.writer.WriteRaw("\x1b[?2004l")
	// Leave alternate screen buffer.
	t.writer.WriteRaw("\x1b[?1049l")
	// Show cursor.
	t.writer.ShowCursor()
	// Reset style.
	t.writer.ResetStyle()
	_ = t.writer.Flush()

	return term.Restore(t.fd, t.oldState)
}

// Size returns the terminal dimensions.
func (t *Terminal) Size() (cols, rows int, err error) {
	ws, err := unix.IoctlGetWinsize(t.fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, fmt.Errorf("term: get winsize: %w", err)
	}
	return int(ws.Col), int(ws.Row), nil
}

// Writer returns the terminal's buffered ANSI writer.
func (t *Terminal) Writer() *Writer { return t.writer }

// Reader returns the terminal's raw byte reader.
func (t *Terminal) Reader() *Reader { return t.reader }

// StartReader begins the background reader goroutine.
func (t *Terminal) StartReader(ctx context.Context) {
	t.reader.Start(ctx)
}
