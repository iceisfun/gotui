package widget

import (
	"github.com/iceisfun/gorepl/pkg/input"
	"github.com/iceisfun/gorepl/pkg/render"
	"github.com/iceisfun/gorepl/pkg/text"
)

// TreeNode represents a node in a tree view.
type TreeNode struct {
	Label    string
	Children []*TreeNode
	Expanded bool
	Data     any
}

// IsLeaf reports whether the node has no children.
func (n *TreeNode) IsLeaf() bool { return len(n.Children) == 0 }

// flatEntry is a flattened tree node with depth and reference.
type flatEntry struct {
	node  *TreeNode
	depth int
}

// Tree is an expandable tree view widget.
// It implements render.Renderable, render.Interactive, render.Focusable, and render.Scrollable.
type Tree struct {
	Roots    []*TreeNode
	OnSelect func(node *TreeNode)

	focused  bool
	cursor   int
	scrollY  int
	flat     []flatEntry
}

// NewTree creates a Tree with the given root nodes.
func NewTree(roots []*TreeNode) *Tree {
	t := &Tree{Roots: roots}
	t.rebuild()
	return t
}

// Focus gives the tree keyboard focus.
func (t *Tree) Focus() { t.focused = true }

// Blur removes keyboard focus from the tree.
func (t *Tree) Blur() { t.focused = false }

// IsFocused reports whether the tree has keyboard focus.
func (t *Tree) IsFocused() bool { return t.focused }

// ContentSize returns the total content dimensions.
func (t *Tree) ContentSize() (w, h int) {
	t.rebuild()
	maxW := 0
	for _, fe := range t.flat {
		lineW := fe.depth*2 + 2 + len([]rune(fe.node.Label))
		if lineW > maxW {
			maxW = lineW
		}
	}
	return maxW, len(t.flat)
}

// ScrollOffset returns the current scroll position.
func (t *Tree) ScrollOffset() (x, y int) { return 0, t.scrollY }

// SetScrollOffset sets the scroll position.
func (t *Tree) SetScrollOffset(x, y int) {
	t.scrollY = y
	if t.scrollY < 0 {
		t.scrollY = 0
	}
}

// Render draws the tree into the view.
func (t *Tree) Render(v *render.View) {
	h := v.Height()
	if h < 1 {
		return
	}

	t.rebuild()
	t.clampScroll(h)

	normalStyle := text.Style{}
	cursorStyle := text.Style{}.Reverse()
	if t.focused {
		cursorStyle = cursorStyle.Bold()
	}
	markerStyle := text.Style{}.WithFg(text.Yellow())

	for row := 0; row < h; row++ {
		idx := row + t.scrollY
		if idx >= len(t.flat) {
			break
		}
		fe := t.flat[idx]
		col := 0

		// Indent.
		indent := fe.depth * 2
		for i := 0; i < indent; i++ {
			v.SetRune(col, row, ' ', normalStyle)
			col++
		}

		// Expand marker.
		var marker rune
		if fe.node.IsLeaf() {
			marker = ' '
		} else if fe.node.Expanded {
			marker = '\u25BC' // down-pointing triangle
		} else {
			marker = '\u25B6' // right-pointing triangle
		}
		v.SetRune(col, row, marker, markerStyle)
		col++
		v.SetRune(col, row, ' ', normalStyle)
		col++

		// Label.
		st := normalStyle
		if idx == t.cursor {
			st = cursorStyle
		}
		v.WriteString(col, row, fe.node.Label, st)
	}
}

// HandleEvent processes keyboard and mouse events.
func (t *Tree) HandleEvent(ev input.Event) bool {
	switch ev.Type {
	case input.EventKey:
		switch ev.Key.Code {
		case input.KeyUp:
			if t.cursor > 0 {
				t.cursor--
			}
			return true
		case input.KeyDown:
			if t.cursor < len(t.flat)-1 {
				t.cursor++
			}
			return true
		case input.KeyRight:
			t.expandCurrent()
			return true
		case input.KeyEnter:
			t.toggleOrSelect()
			return true
		case input.KeyLeft:
			t.collapseCurrent()
			return true
		}
	case input.EventMouse:
		switch ev.Mouse.Button {
		case input.MouseLeft:
			idx := ev.Mouse.Y + t.scrollY
			if idx >= 0 && idx < len(t.flat) {
				t.cursor = idx
				t.toggleOrSelect()
			}
			return true
		case input.MouseWheelUp:
			t.scrollY -= 3
			if t.scrollY < 0 {
				t.scrollY = 0
			}
			return true
		case input.MouseWheelDown:
			t.scrollY += 3
			return true
		}
	}
	return false
}

func (t *Tree) expandCurrent() {
	if t.cursor >= 0 && t.cursor < len(t.flat) {
		node := t.flat[t.cursor].node
		if !node.IsLeaf() && !node.Expanded {
			node.Expanded = true
			t.rebuild()
		}
	}
}

func (t *Tree) collapseCurrent() {
	if t.cursor >= 0 && t.cursor < len(t.flat) {
		node := t.flat[t.cursor].node
		if !node.IsLeaf() && node.Expanded {
			node.Expanded = false
			t.rebuild()
		}
	}
}

func (t *Tree) toggleOrSelect() {
	if t.cursor < 0 || t.cursor >= len(t.flat) {
		return
	}
	node := t.flat[t.cursor].node
	if node.IsLeaf() {
		if t.OnSelect != nil {
			t.OnSelect(node)
		}
	} else {
		node.Expanded = !node.Expanded
		t.rebuild()
	}
}

func (t *Tree) rebuild() {
	t.flat = t.flat[:0]
	for _, root := range t.Roots {
		t.flatten(root, 0)
	}
	if t.cursor >= len(t.flat) {
		t.cursor = len(t.flat) - 1
	}
	if t.cursor < 0 && len(t.flat) > 0 {
		t.cursor = 0
	}
}

func (t *Tree) flatten(node *TreeNode, depth int) {
	t.flat = append(t.flat, flatEntry{node: node, depth: depth})
	if node.Expanded {
		for _, child := range node.Children {
			t.flatten(child, depth+1)
		}
	}
}

func (t *Tree) clampScroll(viewHeight int) {
	if t.cursor < t.scrollY {
		t.scrollY = t.cursor
	}
	if t.cursor >= t.scrollY+viewHeight {
		t.scrollY = t.cursor - viewHeight + 1
	}
	maxScroll := len(t.flat) - viewHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if t.scrollY > maxScroll {
		t.scrollY = maxScroll
	}
	if t.scrollY < 0 {
		t.scrollY = 0
	}
}
