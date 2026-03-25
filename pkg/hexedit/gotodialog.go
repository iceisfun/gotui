package hexedit

import (
	"strconv"
	"strings"

	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// GotoDialog is a modal for jumping to a specific offset.
type GotoDialog struct {
	visible bool
	input   []rune
	cursor  int
	message string

	// OnAction is called when the user submits ("goto") or cancels ("cancel").
	OnAction func(action string)

	borderStyle text.Style
	inputStyle  text.Style
	errorStyle  text.Style
	labelStyle  text.Style
}

// NewGotoDialog creates a goto dialog.
func NewGotoDialog() *GotoDialog {
	return &GotoDialog{
		borderStyle: text.Style{}.WithFg(text.Cyan()),
		inputStyle:  text.Style{}.WithFg(text.White()),
		errorStyle:  text.Style{}.WithFg(text.Red()),
		labelStyle:  text.Style{}.WithFg(text.Cyan()),
	}
}

func (d *GotoDialog) Show()           { d.visible = true; d.input = d.input[:0]; d.cursor = 0; d.message = "" }
func (d *GotoDialog) Hide()           { d.visible = false; d.message = "" }
func (d *GotoDialog) IsVisible() bool { return d.visible }

func (d *GotoDialog) SetMessage(msg string) { d.message = msg }
func (d *GotoDialog) Input() string          { return string(d.input) }

func (d *GotoDialog) Size() (w, h int) { return 40, 5 }

// ParseOffset interprets the input as a byte offset.
// Accepts decimal, 0x hex, and 0o octal prefixes.
func (d *GotoDialog) ParseOffset() (int64, error) {
	s := strings.TrimSpace(string(d.input))
	return strconv.ParseInt(s, 0, 64)
}

func (d *GotoDialog) HandleEvent(ev input.Event) bool {
	if !d.visible || ev.Type != input.EventKey {
		return false
	}

	key := ev.Key
	switch key.Code {
	case input.KeyEscape:
		d.Hide()
		d.fireAction("cancel")
		return true
	case input.KeyEnter:
		d.fireAction("goto")
		return true
	case input.KeyBackspace:
		if d.cursor > 0 {
			d.input = append(d.input[:d.cursor-1], d.input[d.cursor:]...)
			d.cursor--
			d.message = ""
		}
		return true
	case input.KeyDelete:
		if d.cursor < len(d.input) {
			d.input = append(d.input[:d.cursor], d.input[d.cursor+1:]...)
			d.message = ""
		}
		return true
	case input.KeyLeft:
		if d.cursor > 0 {
			d.cursor--
		}
		return true
	case input.KeyRight:
		if d.cursor < len(d.input) {
			d.cursor++
		}
		return true
	case input.KeyHome:
		d.cursor = 0
		return true
	case input.KeyEnd:
		d.cursor = len(d.input)
		return true
	case input.KeyRune:
		if key.Mod == 0 || key.Mod == input.ModShift {
			newInput := make([]rune, len(d.input)+1)
			copy(newInput, d.input[:d.cursor])
			newInput[d.cursor] = key.Rune
			copy(newInput[d.cursor+1:], d.input[d.cursor:])
			d.input = newInput
			d.cursor++
			d.message = ""
			return true
		}
	}

	return false
}

func (d *GotoDialog) fireAction(action string) {
	if d.OnAction != nil {
		d.OnAction(action)
	}
}

func (d *GotoDialog) Render(v *render.View) {
	w, h := v.Width(), v.Height()

	bgStyle := text.Style{}.WithBg(text.Color256(235))
	for y := range h {
		for x := range w {
			v.SetRune(x, y, ' ', bgStyle)
		}
	}

	drawBorder(v, w, h, d.borderStyle, " Go To Offset ", d.borderStyle)

	inner := v.Sub(render.Rect{X: 2, Y: 1, Width: w - 4, Height: h - 2})

	y := 0
	inner.WriteString(0, y, "Offset:", d.labelStyle)
	inputX := 8
	for i, r := range d.input {
		style := d.inputStyle
		if i == d.cursor {
			style = text.Style{}.WithFg(text.Black()).WithBg(text.White())
		}
		inner.SetRune(inputX+i, y, r, style)
	}
	if d.cursor >= len(d.input) {
		inner.SetRune(inputX+len(d.input), y, ' ', text.Style{}.WithBg(text.White()))
	}
	y++

	if d.message != "" {
		inner.WriteString(0, y, d.message, d.errorStyle)
	} else {
		inner.WriteString(0, y, "Dec/0xHex/0oOct, +N/-N relative", text.Style{}.WithFg(text.BrightBlack()))
	}
}
