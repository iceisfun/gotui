package render

import (
	"testing"

	"github.com/iceisfun/gotui/pkg/text"
)

func TestViewClipsWrites(t *testing.T) {
	buf := NewBuffer(20, 20)
	buf.Clear()
	v := NewView(buf, Rect{X: 5, Y: 5, Width: 5, Height: 5})

	// Write inside the view.
	v.SetRune(0, 0, 'A', text.Style{})
	if buf.Cell(5, 5).Rune != 'A' {
		t.Fatal("write inside view should appear in buffer")
	}

	// Write outside the view (should be clipped).
	v.SetRune(5, 0, 'B', text.Style{})
	if buf.Cell(10, 5).Rune == 'B' {
		t.Fatal("write outside view should be clipped")
	}

	v.SetRune(-1, 0, 'C', text.Style{})
	if buf.Cell(4, 5).Rune == 'C' {
		t.Fatal("negative coords should be clipped")
	}
}

func TestViewDimensions(t *testing.T) {
	buf := NewBuffer(20, 20)
	v := NewView(buf, Rect{X: 2, Y: 3, Width: 8, Height: 6})
	if v.Width() != 8 || v.Height() != 6 {
		t.Fatalf("dimensions = %dx%d, want 8x6", v.Width(), v.Height())
	}
}

func TestSubView(t *testing.T) {
	buf := NewBuffer(20, 20)
	buf.Clear()
	v := NewView(buf, Rect{X: 5, Y: 5, Width: 10, Height: 10})
	sub := v.Sub(Rect{X: 2, Y: 2, Width: 3, Height: 3})

	if sub.Width() != 3 || sub.Height() != 3 {
		t.Fatalf("sub dimensions = %dx%d, want 3x3", sub.Width(), sub.Height())
	}

	sub.SetRune(0, 0, 'X', text.Style{})
	if buf.Cell(7, 7).Rune != 'X' {
		t.Fatal("sub-view write should map to correct buffer position")
	}

	// Sub-view should clip to parent.
	sub.SetRune(3, 0, 'Y', text.Style{})
	if buf.Cell(10, 7).Rune == 'Y' {
		t.Fatal("sub-view write outside its bounds should be clipped")
	}
}

func TestSubViewClipsToParent(t *testing.T) {
	buf := NewBuffer(20, 20)
	v := NewView(buf, Rect{X: 5, Y: 5, Width: 5, Height: 5})
	// Sub extends beyond parent.
	sub := v.Sub(Rect{X: 3, Y: 3, Width: 10, Height: 10})
	// Should be clipped to parent bounds.
	if sub.Width() != 2 || sub.Height() != 2 {
		t.Fatalf("sub dimensions = %dx%d, want 2x2", sub.Width(), sub.Height())
	}
}

func TestViewWriteStringTruncation(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.Clear()
	v := NewView(buf, Rect{X: 0, Y: 0, Width: 3, Height: 1})
	n := v.WriteString(0, 0, "hello", text.Style{})
	if n != 3 {
		t.Fatalf("WriteString returned %d, want 3", n)
	}
	// Should only have written 3 chars.
	if buf.Cell(3, 0).Rune == 'l' {
		t.Fatal("WriteString should not write beyond view width")
	}
}

func TestViewWriteStringOutOfBoundsY(t *testing.T) {
	buf := NewBuffer(20, 20)
	v := NewView(buf, Rect{X: 0, Y: 0, Width: 10, Height: 2})
	n := v.WriteString(0, 5, "hello", text.Style{})
	if n != 0 {
		t.Fatalf("WriteString at out-of-bounds y should return 0, got %d", n)
	}
}

func TestViewSetRuneOutOfView(t *testing.T) {
	buf := NewBuffer(20, 20)
	buf.Clear()
	v := NewView(buf, Rect{X: 5, Y: 5, Width: 5, Height: 5})

	// Various out-of-bounds writes — none should panic.
	v.SetRune(-1, -1, 'A', text.Style{})
	v.SetRune(100, 100, 'B', text.Style{})
	v.SetRune(0, 5, 'C', text.Style{})
	v.SetRune(5, 0, 'D', text.Style{})

	// None of those should have written anything outside the view's area.
	if buf.Cell(4, 4).Rune != ' ' {
		t.Fatal("write at (-1,-1) should be silently ignored")
	}
}

func TestViewClear(t *testing.T) {
	buf := NewBuffer(10, 10)
	buf.ClearTransparent()
	v := NewView(buf, Rect{X: 2, Y: 2, Width: 3, Height: 3})
	v.Clear()

	// Inside view should be blank (opaque).
	c := buf.Cell(2, 2)
	if c.Rune != ' ' || c.Transparent {
		t.Fatal("Clear should set opaque blank cells")
	}
	// Outside view should remain transparent.
	if !buf.Cell(0, 0).Transparent {
		t.Fatal("Clear should not affect cells outside the view")
	}
}

func TestViewBounds(t *testing.T) {
	buf := NewBuffer(20, 20)
	bounds := Rect{X: 3, Y: 4, Width: 7, Height: 8}
	v := NewView(buf, bounds)
	if v.Bounds() != bounds {
		t.Fatalf("Bounds() = %+v, want %+v", v.Bounds(), bounds)
	}
}

func TestViewSetCell(t *testing.T) {
	buf := NewBuffer(10, 10)
	buf.Clear()
	v := NewView(buf, Rect{X: 1, Y: 1, Width: 5, Height: 5})
	cell := Cell{Rune: 'Z', Width: 1, Style: text.Style{}.WithFg(text.Blue())}
	v.SetCell(2, 3, cell)
	got := buf.Cell(3, 4)
	if got.Rune != 'Z' {
		t.Fatalf("Rune = %c, want Z", got.Rune)
	}
}
