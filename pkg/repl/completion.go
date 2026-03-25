package repl

import (
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// CompletionPopup renders a list of completion proposals as a floating overlay.
type CompletionPopup struct {
	items    []Completion
	selected int
	visible  bool
}

// NewCompletionPopup creates a completion popup.
func NewCompletionPopup() *CompletionPopup {
	return &CompletionPopup{}
}

// Show displays the popup with the given completions.
func (p *CompletionPopup) Show(items []Completion) {
	p.items = items
	p.selected = 0
	p.visible = len(items) > 0
}

// Hide hides the popup.
func (p *CompletionPopup) Hide() {
	p.visible = false
	p.items = nil
	p.selected = 0
}

// IsVisible reports whether the popup is showing.
func (p *CompletionPopup) IsVisible() bool {
	return p.visible && len(p.items) > 0
}

// Selected returns the currently selected completion, or empty if none.
func (p *CompletionPopup) Selected() (Completion, bool) {
	if !p.IsVisible() || p.selected >= len(p.items) {
		return Completion{}, false
	}
	return p.items[p.selected], true
}

// Next moves selection down.
func (p *CompletionPopup) Next() {
	if !p.IsVisible() {
		return
	}
	p.selected = (p.selected + 1) % len(p.items)
}

// Prev moves selection up.
func (p *CompletionPopup) Prev() {
	if !p.IsVisible() {
		return
	}
	p.selected--
	if p.selected < 0 {
		p.selected = len(p.items) - 1
	}
}

// Size returns the width and height needed for the popup.
func (p *CompletionPopup) Size() (w, h int) {
	if !p.IsVisible() {
		return 0, 0
	}
	maxW := 0
	for _, item := range p.items {
		tw := len(item.Text)
		if item.Detail != "" {
			tw += 2 + len(item.Detail) // "  detail"
		}
		maxW = max(maxW, tw)
	}
	maxH := min(len(p.items), 10) // Cap at 10 visible rows.
	return maxW + 2, maxH         // +2 for padding.
}

// Render draws the popup.
func (p *CompletionPopup) Render(v *render.View) {
	if !p.IsVisible() {
		return
	}

	w, h := v.Width(), v.Height()
	normalStyle := text.Style{}.WithFg(text.White()).WithBg(text.Color256(236))
	selectedStyle := text.Style{}.WithFg(text.Black()).WithBg(text.Cyan())
	detailStyle := text.Style{}.WithFg(text.BrightBlack()).WithBg(text.Color256(236))
	detailSelStyle := text.Style{}.WithFg(text.BrightBlack()).WithBg(text.Cyan())

	// Determine visible window.
	visibleStart := 0
	if p.selected >= h {
		visibleStart = p.selected - h + 1
	}

	for i := 0; i < h && visibleStart+i < len(p.items); i++ {
		idx := visibleStart + i
		item := p.items[idx]
		style := normalStyle
		dstyle := detailStyle
		if idx == p.selected {
			style = selectedStyle
			dstyle = detailSelStyle
		}

		// Fill background.
		for x := 0; x < w; x++ {
			v.SetRune(x, i, ' ', style)
		}

		// Item text.
		col := v.WriteString(1, i, item.Text, style)

		// Detail.
		if item.Detail != "" && col+3 < w {
			v.WriteString(1+col+1, i, item.Detail, dstyle)
		}
	}
}
