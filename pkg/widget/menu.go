package widget

import (
	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/text"
)

// MenuItem represents a single entry in a menu.
type MenuItem struct {
	Label     string
	Shortcut  string
	Action    func()
	Separator bool
}

// Menu is a popup menu widget.
// It implements render.Renderable and render.Interactive.
type Menu struct {
	Items    []MenuItem
	Selected int
	Visible  bool
	OnClose  func()

	// hitRegions maps row index to item index for mouse clicks.
	hitRows []int
}

// NewMenu creates a Menu with the given items.
func NewMenu(items []MenuItem) *Menu {
	return &Menu{Items: items, Selected: firstSelectable(items)}
}

// Show makes the menu visible and resets selection.
func (m *Menu) Show() {
	m.Visible = true
	m.Selected = firstSelectable(m.Items)
}

// Hide hides the menu.
func (m *Menu) Hide() {
	m.Visible = false
}

// Render draws the menu into the view.
func (m *Menu) Render(v *render.View) {
	if !m.Visible || len(m.Items) == 0 {
		return
	}

	// Calculate menu dimensions.
	contentWidth := 0
	for _, item := range m.Items {
		if item.Separator {
			continue
		}
		w := len([]rune(item.Label))
		if item.Shortcut != "" {
			w += 2 + len([]rune(item.Shortcut)) // gap + shortcut
		}
		if w > contentWidth {
			contentWidth = w
		}
	}

	totalRows := len(m.Items)
	boxW := contentWidth + 4 // 2 border + 2 padding
	boxH := totalRows + 2    // 2 border

	borderStyle := text.Style{}.WithFg(text.White())
	normalStyle := text.Style{}
	selectedStyle := text.Style{}.Reverse().Bold()
	sepStyle := text.Style{}.Dim()
	shortcutStyle := text.Style{}.Dim()

	// Top border.
	v.SetRune(0, 0, '\u250C', borderStyle) // top-left corner
	for x := 1; x < boxW-1; x++ {
		v.SetRune(x, 0, '\u2500', borderStyle) // horizontal line
	}
	v.SetRune(boxW-1, 0, '\u2510', borderStyle) // top-right corner

	m.hitRows = make([]int, 0, totalRows)

	for i, item := range m.Items {
		row := i + 1
		if row >= boxH-1 || row >= v.Height() {
			break
		}

		v.SetRune(0, row, '\u2502', borderStyle) // left border

		if item.Separator {
			for x := 1; x < boxW-1; x++ {
				v.SetRune(x, row, '\u2500', sepStyle)
			}
			m.hitRows = append(m.hitRows, -1)
		} else {
			st := normalStyle
			if i == m.Selected {
				st = selectedStyle
			}
			// Fill interior with spaces.
			for x := 1; x < boxW-1; x++ {
				v.SetRune(x, row, ' ', st)
			}
			// Label.
			v.WriteString(2, row, item.Label, st)
			// Right-aligned shortcut.
			if item.Shortcut != "" {
				scStyle := shortcutStyle
				if i == m.Selected {
					scStyle = st
				}
				scX := boxW - 2 - len([]rune(item.Shortcut))
				v.WriteString(scX, row, item.Shortcut, scStyle)
			}
			m.hitRows = append(m.hitRows, i)
		}

		v.SetRune(boxW-1, row, '\u2502', borderStyle) // right border
	}

	// Bottom border.
	bottomRow := totalRows + 1
	if bottomRow < v.Height() {
		v.SetRune(0, bottomRow, '\u2514', borderStyle) // bottom-left corner
		for x := 1; x < boxW-1; x++ {
			v.SetRune(x, bottomRow, '\u2500', borderStyle)
		}
		v.SetRune(boxW-1, bottomRow, '\u2518', borderStyle) // bottom-right corner
	}
}

// HandleEvent processes keyboard and mouse events.
func (m *Menu) HandleEvent(ev input.Event) bool {
	if !m.Visible {
		return false
	}

	switch ev.Type {
	case input.EventKey:
		switch ev.Key.Code {
		case input.KeyUp:
			m.movePrev()
			return true
		case input.KeyDown:
			m.moveNext()
			return true
		case input.KeyEnter:
			m.activate()
			return true
		case input.KeyEscape:
			m.Hide()
			if m.OnClose != nil {
				m.OnClose()
			}
			return true
		}
	case input.EventMouse:
		if ev.Mouse.Button == input.MouseLeft {
			row := ev.Mouse.Y - 1 // subtract top border
			if row >= 0 && row < len(m.hitRows) && m.hitRows[row] >= 0 {
				m.Selected = m.hitRows[row]
				m.activate()
				return true
			}
		}
	}
	return false
}

func (m *Menu) activate() {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		item := m.Items[m.Selected]
		if !item.Separator && item.Action != nil {
			m.Hide()
			item.Action()
		}
	}
}

func (m *Menu) moveNext() {
	for i := m.Selected + 1; i < len(m.Items); i++ {
		if !m.Items[i].Separator {
			m.Selected = i
			return
		}
	}
}

func (m *Menu) movePrev() {
	for i := m.Selected - 1; i >= 0; i-- {
		if !m.Items[i].Separator {
			m.Selected = i
			return
		}
	}
}

func firstSelectable(items []MenuItem) int {
	for i, item := range items {
		if !item.Separator {
			return i
		}
	}
	return 0
}
