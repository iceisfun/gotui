package hexedit

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/text"
)

// SearchDialog is a modal search box with mode selection and alignment.
type SearchDialog struct {
	visible bool
	mode    SearchMode
	align   Alignment
	input   []rune
	cursor  int
	message string // Status/error message.

	// OnAction is called when the user submits ("search") or cancels ("cancel").
	OnAction func(action string)

	borderStyle text.Style
	inputStyle  text.Style
	modeStyle   text.Style
	activeStyle text.Style
	errorStyle  text.Style
	labelStyle  text.Style

	// Hit regions recorded during Render for mouse click handling.
	modeHits  []hitRegion[SearchMode]
	alignHits []hitRegion[Alignment]
}

type hitRegion[T any] struct {
	x, w int // Column start and width on the row.
	row  int
	val  T
}

// NewSearchDialog creates a search dialog.
func NewSearchDialog() *SearchDialog {
	return &SearchDialog{
		borderStyle: text.Style{}.WithFg(text.Cyan()),
		inputStyle:  text.Style{}.WithFg(text.White()),
		modeStyle:   text.Style{}.WithFg(text.BrightBlack()),
		activeStyle: text.Style{}.WithFg(text.Black()).WithBg(text.Cyan()),
		errorStyle:  text.Style{}.WithFg(text.Red()),
		labelStyle:  text.Style{}.WithFg(text.Cyan()),
	}
}

// Show opens the search dialog.
func (d *SearchDialog) Show() {
	d.visible = true
	d.input = d.input[:0]
	d.cursor = 0
	d.message = ""
}

// Hide closes the search dialog.
func (d *SearchDialog) Hide() {
	d.visible = false
	d.message = ""
}

// IsVisible reports whether the dialog is showing.
func (d *SearchDialog) IsVisible() bool { return d.visible }

// Mode returns the current search mode.
func (d *SearchDialog) Mode() SearchMode { return d.mode }

// Align returns the current alignment constraint.
func (d *SearchDialog) Align() Alignment { return d.align }

// Input returns the current search input.
func (d *SearchDialog) Input() string { return string(d.input) }

// SetMessage sets a status message (e.g., "Not found").
func (d *SearchDialog) SetMessage(msg string) { d.message = msg }

// Size returns the dialog dimensions.
func (d *SearchDialog) Size() (w, h int) {
	return 55, 8
}

// HandleEvent processes input for the dialog. Implements render.Interactive.
// Actions ("search", "cancel") are delivered via the OnAction callback.
func (d *SearchDialog) HandleEvent(ev input.Event) bool {
	if !d.visible {
		return false
	}

	// Mouse clicks on mode/alignment buttons.
	if ev.Type == input.EventMouse && ev.Mouse.Button == input.MouseLeft {
		// Coords are local to overlay. Inner area is offset by border (2,1).
		mx, my := ev.Mouse.X-2, ev.Mouse.Y-1
		for _, h := range d.modeHits {
			if my == h.row && mx >= h.x && mx < h.x+h.w {
				d.mode = h.val
				d.message = ""
				return true
			}
		}
		for _, h := range d.alignHits {
			if my == h.row && mx >= h.x && mx < h.x+h.w {
				d.align = h.val
				d.message = ""
				return true
			}
		}
		return true
	}

	if ev.Type != input.EventKey {
		return false
	}

	key := ev.Key

	switch key.Code {
	case input.KeyEscape:
		d.Hide()
		d.fireAction("cancel")
		return true

	case input.KeyEnter:
		d.fireAction("search")
		return true

	case input.KeyTab:
		d.mode = (d.mode + 1) % (SearchU64 + 1)
		d.message = ""
		return true

	case input.KeyBacktab:
		d.mode--
		if d.mode < 0 {
			d.mode = SearchU64
		}
		d.message = ""
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
		if key.Mod == input.ModAlt && (key.Rune == 'a' || key.Rune == 'A') {
			d.align = NextAlignment(d.align)
			d.message = ""
			return true
		}
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

func (d *SearchDialog) fireAction(action string) {
	if d.OnAction != nil {
		d.OnAction(action)
	}
}

// Render draws the search dialog.
func (d *SearchDialog) Render(v *render.View) {
	w, h := v.Width(), v.Height()

	// Background fill.
	bgStyle := text.Style{}.WithBg(text.Color256(235))
	for y := range h {
		for x := range w {
			v.SetRune(x, y, ' ', bgStyle)
		}
	}

	drawBorder(v, w, h, d.borderStyle, " Search ", d.borderStyle)

	inner := v.Sub(render.Rect{X: 2, Y: 1, Width: w - 4, Height: h - 2})

	// Mode selector (Tab to cycle, click to select).
	y := 0
	d.modeHits = d.modeHits[:0]
	inner.WriteString(0, y, "Mode:", d.labelStyle)
	x := 6
	for i := SearchString; i <= SearchU64; i++ {
		name := "[" + i.String() + "]"
		style := d.modeStyle
		if i == d.mode {
			style = d.activeStyle
		}
		d.modeHits = append(d.modeHits, hitRegion[SearchMode]{x: x, w: len(name), row: y, val: i})
		x += inner.WriteString(x, y, name, style)
		x += inner.WriteString(x, y, " ", bgStyle)
	}
	y++

	// Alignment selector (Alt-A to cycle, click to select).
	d.alignHits = d.alignHits[:0]
	inner.WriteString(0, y, "Align:", d.labelStyle)
	ax := 7
	for _, a := range []Alignment{AlignNone, Align32, Align64} {
		name := "[" + a.String() + "]"
		style := d.modeStyle
		if a == d.align {
			style = d.activeStyle
		}
		d.alignHits = append(d.alignHits, hitRegion[Alignment]{x: ax, w: len(name), row: y, val: a})
		ax += inner.WriteString(ax, y, name, style)
		ax += inner.WriteString(ax, y, " ", bgStyle)
	}
	y++

	// Input field.
	inner.WriteString(0, y, "Find:", d.labelStyle)
	inputX := 6
	for i, r := range d.input {
		style := d.inputStyle
		if i == d.cursor {
			style = text.Style{}.WithFg(text.Black()).WithBg(text.White())
		}
		inner.SetRune(inputX+i, y, r, style)
	}
	// Show cursor at end.
	if d.cursor >= len(d.input) {
		inner.SetRune(inputX+len(d.input), y, ' ', text.Style{}.WithBg(text.White()))
	}
	y++

	// Preview.
	if len(d.input) > 0 {
		pattern, err := ParseSearchPattern(string(d.input), d.mode)
		if err == nil && len(pattern) > 0 {
			preview := FormatSearchPreview(pattern)
			inner.WriteString(0, y, "Bytes:", d.labelStyle)
			inner.WriteString(7, y, preview, text.Style{}.WithFg(text.Yellow()))
		}
	}
	y++

	// Message.
	if d.message != "" {
		inner.WriteString(0, y, d.message, d.errorStyle)
	} else {
		hint := "Enter=search  Tab=mode  Alt-A=align  Esc=cancel"
		inner.WriteString(0, y, hint, text.Style{}.WithFg(text.BrightBlack()))
	}
}
