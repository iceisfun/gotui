package render

import "github.com/iceisfun/gotui/pkg/text"

// Cell is one character position in the framebuffer.
type Cell struct {
	Rune        rune
	Width       uint8      // Display width: 1 for ASCII, 2 for CJK, 0 for continuation.
	Style       text.Style
	Transparent bool       // True = falls through to the layer below during compositing.
}

// BlankCell is a visible space with default style.
var BlankCell = Cell{Rune: ' ', Width: 1}

// TransparentCell is a cell that falls through during compositing.
var TransparentCell = Cell{Transparent: true}
