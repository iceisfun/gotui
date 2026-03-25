package repl

import (
	"strings"

	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/text"
)

// Option configures the REPL.
type Option func(*REPL)

func WithCompleter(c Completer) Option     { return func(r *REPL) { r.completer = c } }
func WithHighlighter(h Highlighter) Option { return func(r *REPL) { r.highlighter = h } }
func WithPrompt(primary, continuation string) Option {
	return func(r *REPL) {
		r.prompt = primary
		r.promptCont = continuation
	}
}
func WithHistorySize(n int) Option { return func(r *REPL) { r.history = NewHistory(n) } }

// REPL is the main interactive component. It implements:
// Renderable, Interactive, Focusable, Cursorable, Scrollable, Overlayable.
type REPL struct {
	editor      *Editor
	history     *History
	output      *Output
	completion  *CompletionPopup
	executor    Executor
	completer   Completer
	highlighter Highlighter

	prompt     string
	promptCont string

	focused bool
	scrollY int // Scroll offset for the entire REPL view.

	// Dimensions from last render.
	viewWidth  int
	viewHeight int

	// Prompt style.
	promptStyle text.Style
	errorStyle  text.Style
	resultStyle text.Style
}

// New creates a REPL with the given executor and options.
func New(executor Executor, opts ...Option) *REPL {
	r := &REPL{
		editor:      NewEditor(),
		history:     NewHistory(1000),
		output:      NewOutput(),
		completion:  NewCompletionPopup(),
		executor:    executor,
		prompt:      "> ",
		promptCont:  "  ",
		promptStyle: text.Style{}.WithFg(text.Cyan()).Bold(),
		errorStyle:  text.Style{}.WithFg(text.Red()),
		resultStyle: text.Style{}.WithFg(text.Green()),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Render draws the output area, prompt, and editor content.
func (r *REPL) Render(v *render.View) {
	r.viewWidth = v.Width()
	r.viewHeight = v.Height()

	y := 0

	// Render output lines (above the editor).
	outputLines := r.output.Lines()
	editorLines := r.editor.Lines()
	promptLen := len([]rune(r.prompt))

	// How many lines the editor occupies.
	editorHeight := editorLines

	// Available space for output.
	outputSpace := r.viewHeight - editorHeight
	if outputSpace < 0 {
		outputSpace = 0
	}

	// Render output, showing the tail.
	if outputLines > 0 && outputSpace > 0 {
		visibleCount := min(outputLines, outputSpace)
		startIdx := outputLines - visibleCount
		for i := startIdx; i < outputLines && y < outputSpace; i++ {
			line := r.output.lines[i]
			col := 0
			for _, span := range line {
				col += v.WriteString(col, y, span.Text, span.Style)
			}
			y++
		}
	}

	// Render editor lines with prompts and optional syntax highlighting.
	for i := 0; i < editorLines && y < r.viewHeight; i++ {
		// Prompt.
		prompt := r.prompt
		if i > 0 {
			prompt = r.promptCont
		}
		v.WriteString(0, y, prompt, r.promptStyle)

		// Line content.
		lineStr := r.editor.Line(i)
		if r.highlighter != nil && lineStr != "" {
			styled := r.highlighter.Highlight(lineStr)
			col := promptLen
			for _, span := range styled {
				col += v.WriteString(col, y, span.Text, span.Style)
			}
		} else {
			v.WriteString(promptLen, y, lineStr, text.Style{})
		}

		y++
	}
}

// HandleEvent processes keyboard input.
func (r *REPL) HandleEvent(ev input.Event) bool {
	if ev.Type != input.EventKey {
		return false
	}

	key := ev.Key

	// Completion navigation.
	if r.completion.IsVisible() {
		switch key.Code {
		case input.KeyTab:
			r.completion.Next()
			return true
		case input.KeyBacktab:
			r.completion.Prev()
			return true
		case input.KeyEnter:
			if comp, ok := r.completion.Selected(); ok {
				r.applyCompletion(comp)
			}
			r.completion.Hide()
			return true
		case input.KeyEscape:
			r.completion.Hide()
			return true
		}
		// Any other key hides completion and falls through.
		r.completion.Hide()
	}

	// Ctrl bindings.
	if key.Mod&input.ModCtrl != 0 {
		switch key.Rune {
		case 'a':
			r.editor.MoveHome()
			return true
		case 'e':
			r.editor.MoveEnd()
			return true
		case 'k':
			r.editor.DeleteToEnd()
			return true
		case 'u':
			r.editor.DeleteToStart()
			return true
		case 'w':
			r.editor.DeleteWordBack()
			return true
		case 'l':
			r.output.Clear()
			return true
		case 'p':
			r.historyPrev()
			return true
		case 'n':
			r.historyNext()
			return true
		}
	}

	// Alt bindings.
	if key.Mod&input.ModAlt != 0 {
		switch key.Rune {
		case 'b':
			r.editor.MoveWordLeft()
			return true
		case 'f':
			r.editor.MoveWordRight()
			return true
		}
	}

	switch key.Code {
	case input.KeyEnter:
		r.handleEnter()
		return true

	case input.KeyTab:
		r.handleTab()
		return true

	case input.KeyBackspace:
		r.editor.Backspace()
		return true

	case input.KeyDelete:
		r.editor.Delete()
		return true

	case input.KeyLeft:
		if key.Mod&input.ModCtrl != 0 || key.Mod&input.ModAlt != 0 {
			r.editor.MoveWordLeft()
		} else {
			r.editor.MoveLeft()
		}
		return true

	case input.KeyRight:
		if key.Mod&input.ModCtrl != 0 || key.Mod&input.ModAlt != 0 {
			r.editor.MoveWordRight()
		} else {
			r.editor.MoveRight()
		}
		return true

	case input.KeyUp:
		if r.editor.Lines() == 1 || key.Mod&input.ModAlt != 0 {
			r.historyPrev()
		} else {
			r.editor.MoveUp()
		}
		return true

	case input.KeyDown:
		if r.editor.Lines() == 1 || key.Mod&input.ModAlt != 0 {
			r.historyNext()
		} else {
			r.editor.MoveDown()
		}
		return true

	case input.KeyHome:
		r.editor.MoveHome()
		return true

	case input.KeyEnd:
		r.editor.MoveEnd()
		return true

	case input.KeyRune:
		r.editor.InsertRune(key.Rune)
		return true

	case input.KeyEscape:
		return false // Let it bubble.
	}

	return false
}

func (r *REPL) handleEnter() {
	src := r.editor.Text()

	// Check if the executor considers this complete.
	if r.executor != nil && !r.executor.IsComplete(src) {
		r.editor.InsertNewline()
		return
	}

	// Add input to output display.
	for i := 0; i < r.editor.Lines(); i++ {
		prompt := r.prompt
		if i > 0 {
			prompt = r.promptCont
		}
		line := text.StyledLine{
			{Text: prompt, Style: r.promptStyle},
		}
		lineStr := r.editor.Line(i)
		if r.highlighter != nil && lineStr != "" {
			line = append(line, r.highlighter.Highlight(lineStr)...)
		} else {
			line = append(line, text.StyledSpan{Text: lineStr})
		}
		r.output.Append(line)
	}

	// Execute.
	if r.executor != nil && strings.TrimSpace(src) != "" {
		result, err := r.executor.Execute(src)
		if err != nil {
			r.output.Append(text.Styled(err.Error(), r.errorStyle))
		} else if result != "" {
			// Split result into lines.
			for _, line := range strings.Split(result, "\n") {
				r.output.Append(text.Styled(line, r.resultStyle))
			}
		}
	}

	// Add to history.
	if strings.TrimSpace(src) != "" {
		r.history.Add(src)
	}

	r.editor.Clear()
}

func (r *REPL) handleTab() {
	if r.completer == nil {
		// No completer — insert literal tab (as spaces).
		r.editor.InsertString("  ")
		return
	}

	lineStr := r.editor.Line(r.editor.curRow)
	items := r.completer.Complete(lineStr, r.editor.curCol)
	if len(items) == 0 {
		return
	}
	if len(items) == 1 {
		r.applyCompletion(items[0])
		return
	}
	r.completion.Show(items)
}

func (r *REPL) applyCompletion(comp Completion) {
	// Find the word start before cursor to determine replacement range.
	line := []rune(r.editor.Line(r.editor.curRow))
	col := r.editor.clampCol()

	// Walk back to find word start.
	wordStart := col
	for wordStart > 0 && isIdentChar(line[wordStart-1]) {
		wordStart--
	}

	// Replace the word with the completion text.
	newLine := make([]rune, 0, wordStart+len([]rune(comp.Text))+len(line)-col)
	newLine = append(newLine, line[:wordStart]...)
	compRunes := []rune(comp.Text)
	newLine = append(newLine, compRunes...)
	newLine = append(newLine, line[col:]...)
	r.editor.lines[r.editor.curRow] = newLine
	r.editor.curCol = wordStart + len(compRunes)
}

func isIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.'
}

func (r *REPL) historyPrev() {
	if entry, ok := r.history.Prev(r.editor.Text()); ok {
		r.editor.SetText(entry)
	}
}

func (r *REPL) historyNext() {
	if entry, ok := r.history.Next(r.editor.Text()); ok {
		r.editor.SetText(entry)
	}
}

// Focus / Blur / IsFocused — Focusable interface.
func (r *REPL) Focus()          { r.focused = true }
func (r *REPL) Blur()           { r.focused = false; r.completion.Hide() }
func (r *REPL) IsFocused() bool { return r.focused }

// CursorPosition — Cursorable interface.
func (r *REPL) CursorPosition() (x, y int, visible bool) {
	if !r.focused {
		return 0, 0, false
	}
	row, col := r.editor.CursorPos()

	promptLen := len([]rune(r.prompt))
	if row > 0 {
		promptLen = len([]rune(r.promptCont))
	}

	// y = output lines + editor row.
	outputVisible := r.viewHeight - r.editor.Lines()
	if outputVisible < 0 {
		outputVisible = 0
	}
	outputShown := min(r.output.Lines(), outputVisible)

	return promptLen + col, outputShown + row, true
}

// ContentSize — Scrollable interface.
func (r *REPL) ContentSize() (w, h int) {
	return r.viewWidth, r.output.Lines() + r.editor.Lines()
}

// ScrollOffset — Scrollable interface.
func (r *REPL) ScrollOffset() (x, y int) {
	return 0, r.scrollY
}

// SetScrollOffset — Scrollable interface.
func (r *REPL) SetScrollOffset(x, y int) {
	r.scrollY = max(y, 0)
}

// Overlays — Overlayable interface.
func (r *REPL) Overlays() []render.OverlayRequest {
	if !r.completion.IsVisible() {
		return nil
	}

	w, h := r.completion.Size()
	cx, cy, _ := r.CursorPosition()

	return []render.OverlayRequest{{
		Renderable: r.completion,
		Anchor: render.Rect{
			X:      cx,
			Y:      cy + 1, // Below cursor line.
			Width:  w,
			Height: h,
		},
		ZOrder: 100,
	}}
}
