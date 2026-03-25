package hexedit

import (
	"fmt"

	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// InfoPanel displays byte interpretations at the current cursor position.
type InfoPanel struct {
	hexView *HexView

	borderStyle text.Style
	labelStyle  text.Style
	valueStyle  text.Style
	headerStyle text.Style
}

// NewInfoPanel creates an info panel linked to a hex view.
func NewInfoPanel(hv *HexView) *InfoPanel {
	return &InfoPanel{
		hexView:     hv,
		borderStyle: text.Style{}.WithFg(text.BrightBlack()),
		labelStyle:  text.Style{}.WithFg(text.Cyan()),
		valueStyle:  text.Style{}.WithFg(text.White()),
		headerStyle: text.Style{}.WithFg(text.Cyan()).Bold(),
	}
}

// Render draws the info panel.
func (p *InfoPanel) Render(v *render.View) {
	w, h := v.Width(), v.Height()
	if w < 10 || h < 4 {
		return
	}

	// Border.
	drawBorder(v, w, h, p.borderStyle, " Inspector ", p.headerStyle)

	inner := v.Sub(render.Rect{X: 2, Y: 1, Width: w - 4, Height: h - 2})
	iw := inner.Width()

	cursor := p.hexView.Cursor()
	buf := p.hexView.Buffer()
	interp := buf.InterpretAt(cursor)

	y := 0

	// Position info.
	inner.WriteString(0, y, "Offset:", p.labelStyle)
	inner.WriteString(9, y, fmt.Sprintf("0x%08X (%d)", cursor, cursor), p.valueStyle)
	y++

	remaining := buf.Len() - cursor
	inner.WriteString(0, y, "Remain:", p.labelStyle)
	inner.WriteString(9, y, fmt.Sprintf("%d bytes", remaining), p.valueStyle)
	y++

	// Separator.
	y++
	for x := range iw {
		inner.SetRune(x, y, '─', p.borderStyle)
	}
	y++

	// Signed / Unsigned header.
	const colType = 0
	const colUnsigned = 11
	const colSigned = 33

	inner.WriteString(colType, y, "Type", p.headerStyle)
	inner.WriteString(colUnsigned, y, "Unsigned", p.headerStyle)
	inner.WriteString(colSigned, y, "Signed", p.headerStyle)
	y++

	for x := range iw {
		inner.SetRune(x, y, '─', p.borderStyle)
	}
	y++

	if interp.HasU8 {
		inner.WriteString(colType, y, "uint8", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U8), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I8), p.valueStyle)
		y++
	}

	if interp.HasU16 {
		inner.WriteString(colType, y, "uint16 LE", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U16LE), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I16LE), p.valueStyle)
		y++
		inner.WriteString(colType, y, "uint16 BE", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U16BE), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I16BE), p.valueStyle)
		y++
	}

	if interp.HasU32 {
		inner.WriteString(colType, y, "uint32 LE", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U32LE), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I32LE), p.valueStyle)
		y++
		inner.WriteString(colType, y, "uint32 BE", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U32BE), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I32BE), p.valueStyle)
		y++
	}

	if interp.HasU64 {
		inner.WriteString(colType, y, "uint64 LE", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U64LE), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I64LE), p.valueStyle)
		y++
		inner.WriteString(colType, y, "uint64 BE", p.labelStyle)
		inner.WriteString(colUnsigned, y, fmt.Sprintf("%-21d", interp.U64BE), p.valueStyle)
		inner.WriteString(colSigned, y, fmt.Sprintf("%d", interp.I64BE), p.valueStyle)
		y++
	}

	// Hex representation of next 8 bytes.
	y++
	for x := range iw {
		inner.SetRune(x, y, '─', p.borderStyle)
	}
	y++

	inner.WriteString(0, y, "Hex:", p.labelStyle)
	bytes := buf.Bytes(cursor, 8)
	hexStr := ""
	for i, b := range bytes {
		if i > 0 {
			hexStr += " "
		}
		hexStr += fmt.Sprintf("%02X", b)
	}
	inner.WriteString(10, y, hexStr, p.valueStyle)
	y++

	// Binary representation of current byte.
	if interp.HasU8 {
		inner.WriteString(0, y, "Binary:", p.labelStyle)
		inner.WriteString(10, y, fmt.Sprintf("%08b", interp.U8), p.valueStyle)
		y++
	}

	// ASCII character.
	inner.WriteString(0, y, "Char:", p.labelStyle)
	if interp.HasU8 && interp.U8 >= 0x20 && interp.U8 <= 0x7e {
		inner.WriteString(10, y, fmt.Sprintf("'%c'", interp.U8), p.valueStyle)
	} else if interp.HasU8 {
		inner.WriteString(10, y, fmt.Sprintf("\\x%02x", interp.U8), text.Style{}.WithFg(text.BrightBlack()))
	}
}

func drawBorder(v *render.View, w, h int, style text.Style, title string, titleStyle text.Style) {
	v.SetRune(0, 0, '┌', style)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, 0, '─', style)
	}
	v.SetRune(w-1, 0, '┐', style)

	if title != "" && len(title)+2 < w {
		v.WriteString(2, 0, title, titleStyle)
	}

	for y := 1; y < h-1; y++ {
		v.SetRune(0, y, '│', style)
		v.SetRune(w-1, y, '│', style)
	}

	v.SetRune(0, h-1, '└', style)
	for x := 1; x < w-1; x++ {
		v.SetRune(x, h-1, '─', style)
	}
	v.SetRune(w-1, h-1, '┘', style)
}
