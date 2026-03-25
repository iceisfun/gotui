package render

import (
	"testing"

	"github.com/iceisfun/gotui/pkg/text"
)

func TestNewBufferDimensions(t *testing.T) {
	b := NewBuffer(80, 24)
	if b.Width() != 80 || b.Height() != 24 {
		t.Fatalf("dimensions = %dx%d, want 80x24", b.Width(), b.Height())
	}
}

func TestNewBufferTransparent(t *testing.T) {
	b := NewBuffer(3, 3)
	c := b.Cell(0, 0)
	if !c.Transparent {
		t.Fatal("new buffer cells should be transparent")
	}
}

func TestSetCellAndCell(t *testing.T) {
	b := NewBuffer(10, 10)
	cell := Cell{Rune: 'X', Width: 1, Style: text.Style{}.WithFg(text.Red())}
	b.SetCell(5, 5, cell)
	got := b.Cell(5, 5)
	if got.Rune != 'X' {
		t.Fatalf("Rune = %c, want X", got.Rune)
	}
	if !got.Style.Fg.Equal(text.Red()) {
		t.Fatal("style mismatch")
	}
}

func TestSetRune(t *testing.T) {
	b := NewBuffer(10, 10)
	s := text.Style{}.Bold()
	b.SetRune(3, 4, 'A', s)
	c := b.Cell(3, 4)
	if c.Rune != 'A' || c.Width != 1 {
		t.Fatalf("cell = %+v", c)
	}
	if !c.Style.Equal(s) {
		t.Fatal("style mismatch")
	}
}

func TestWriteString(t *testing.T) {
	b := NewBuffer(10, 1)
	s := text.Style{}
	n := b.WriteString(0, 0, "hello", s)
	if n != 5 {
		t.Fatalf("WriteString returned %d, want 5", n)
	}
	for i, r := range "hello" {
		c := b.Cell(i, 0)
		if c.Rune != r {
			t.Fatalf("cell[%d] = %c, want %c", i, c.Rune, r)
		}
	}
}

func TestWriteStringTruncation(t *testing.T) {
	b := NewBuffer(3, 1)
	n := b.WriteString(0, 0, "hello", text.Style{})
	if n != 3 {
		t.Fatalf("WriteString returned %d, want 3", n)
	}
}

func TestOutOfBounds(t *testing.T) {
	b := NewBuffer(5, 5)
	// Should not panic.
	b.SetCell(-1, 0, BlankCell)
	b.SetCell(0, -1, BlankCell)
	b.SetCell(5, 0, BlankCell)
	b.SetCell(0, 5, BlankCell)
	b.SetRune(100, 100, 'X', text.Style{})

	c := b.Cell(-1, 0)
	if !c.Transparent {
		t.Fatal("out of bounds Cell should return TransparentCell")
	}
	c = b.Cell(5, 5)
	if !c.Transparent {
		t.Fatal("out of bounds Cell should return TransparentCell")
	}
}

func TestClear(t *testing.T) {
	b := NewBuffer(3, 3)
	b.SetRune(0, 0, 'X', text.Style{})
	b.Clear()
	c := b.Cell(0, 0)
	if c.Rune != ' ' || c.Transparent {
		t.Fatalf("after Clear, cell = %+v, want blank opaque", c)
	}
}

func TestClearTransparent(t *testing.T) {
	b := NewBuffer(3, 3)
	b.SetRune(0, 0, 'X', text.Style{})
	b.ClearTransparent()
	c := b.Cell(0, 0)
	if !c.Transparent {
		t.Fatal("after ClearTransparent, cell should be transparent")
	}
}

func TestFill(t *testing.T) {
	b := NewBuffer(5, 5)
	cell := Cell{Rune: '#', Width: 1}
	b.Fill(Rect{X: 1, Y: 1, Width: 3, Height: 3}, cell)

	// Inside the fill region.
	if b.Cell(1, 1).Rune != '#' {
		t.Fatal("inside fill should be '#'")
	}
	if b.Cell(3, 3).Rune != '#' {
		t.Fatal("inside fill should be '#'")
	}
	// Outside the fill region.
	if b.Cell(0, 0).Rune == '#' {
		t.Fatal("outside fill should not be '#'")
	}
}

func TestFillClipsToBuffer(t *testing.T) {
	b := NewBuffer(5, 5)
	cell := Cell{Rune: '#', Width: 1}
	// Rect extends beyond buffer — should not panic.
	b.Fill(Rect{X: 3, Y: 3, Width: 10, Height: 10}, cell)
	if b.Cell(4, 4).Rune != '#' {
		t.Fatal("fill should reach (4,4)")
	}
}

func TestResize(t *testing.T) {
	b := NewBuffer(5, 5)
	b.SetRune(0, 0, 'X', text.Style{})
	b.Resize(10, 10)
	if b.Width() != 10 || b.Height() != 10 {
		t.Fatalf("after resize: %dx%d", b.Width(), b.Height())
	}
	// Old content discarded.
	c := b.Cell(0, 0)
	if !c.Transparent {
		t.Fatal("after resize, cells should be transparent")
	}
}

func TestResizeSameDimensions(t *testing.T) {
	b := NewBuffer(5, 5)
	b.SetRune(0, 0, 'X', text.Style{})
	b.Resize(5, 5)
	// Same dimensions — content preserved.
	if b.Cell(0, 0).Rune != 'X' {
		t.Fatal("same-size resize should preserve content")
	}
}

func TestClearWithStyle(t *testing.T) {
	b := NewBuffer(3, 3)
	s := text.Style{}.WithFg(text.Red())
	b.ClearWithStyle(s)
	c := b.Cell(1, 1)
	if c.Rune != ' ' || c.Transparent {
		t.Fatal("ClearWithStyle should set opaque space cells")
	}
	if !c.Style.Equal(s) {
		t.Fatal("ClearWithStyle should apply the given style")
	}
}
