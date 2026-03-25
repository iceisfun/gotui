package render

import (
	"testing"

	"github.com/iceisfun/gotui/pkg/text"
)

func TestCompositeTopLayerWins(t *testing.T) {
	dst := NewBuffer(3, 1)
	bottom := NewBuffer(3, 1)
	top := NewBuffer(3, 1)

	bottom.SetRune(0, 0, 'A', text.Style{})
	bottom.SetRune(1, 0, 'B', text.Style{})
	bottom.SetRune(2, 0, 'C', text.Style{})

	top.SetRune(1, 0, 'X', text.Style{})
	// top[0] and top[2] are transparent.

	Composite(dst, []*Buffer{bottom, top})

	if dst.Cell(0, 0).Rune != 'A' {
		t.Fatal("transparent top cell should fall through to bottom")
	}
	if dst.Cell(1, 0).Rune != 'X' {
		t.Fatal("opaque top cell should win")
	}
	if dst.Cell(2, 0).Rune != 'C' {
		t.Fatal("transparent top cell should fall through to bottom")
	}
}

func TestCompositeAllTransparent(t *testing.T) {
	dst := NewBuffer(2, 1)
	layer1 := NewBuffer(2, 1) // all transparent
	layer2 := NewBuffer(2, 1) // all transparent

	Composite(dst, []*Buffer{layer1, layer2})

	c := dst.Cell(0, 0)
	if c.Rune != BlankCell.Rune || c.Transparent != BlankCell.Transparent {
		t.Fatalf("all-transparent should fall back to BlankCell, got %+v", c)
	}
}

func TestCompositeNoLayers(t *testing.T) {
	dst := NewBuffer(2, 1)
	dst.SetRune(0, 0, 'X', text.Style{})
	Composite(dst, nil)
	// dst should be unchanged.
	if dst.Cell(0, 0).Rune != 'X' {
		t.Fatal("Composite with no layers should not modify dst")
	}
}

func TestCompositeTwoOverlay(t *testing.T) {
	dst := NewBuffer(3, 1)
	base := NewBuffer(3, 1)
	overlay := NewBuffer(3, 1)

	base.SetRune(0, 0, 'A', text.Style{})
	base.SetRune(1, 0, 'B', text.Style{})
	base.SetRune(2, 0, 'C', text.Style{})

	overlay.SetRune(1, 0, 'Y', text.Style{})

	CompositeTwo(dst, base, overlay)

	if dst.Cell(0, 0).Rune != 'A' {
		t.Fatal("transparent overlay should show base")
	}
	if dst.Cell(1, 0).Rune != 'Y' {
		t.Fatal("opaque overlay should win")
	}
	if dst.Cell(2, 0).Rune != 'C' {
		t.Fatal("transparent overlay should show base")
	}
}

func TestCompositeTwoAllTransparentOverlay(t *testing.T) {
	dst := NewBuffer(2, 1)
	base := NewBuffer(2, 1)
	overlay := NewBuffer(2, 1) // all transparent

	base.SetRune(0, 0, 'Z', text.Style{})

	CompositeTwo(dst, base, overlay)
	if dst.Cell(0, 0).Rune != 'Z' {
		t.Fatal("fully transparent overlay should pass through base")
	}
}

func TestCompositeThreeLayers(t *testing.T) {
	dst := NewBuffer(1, 1)
	l1 := NewBuffer(1, 1)
	l2 := NewBuffer(1, 1)
	l3 := NewBuffer(1, 1)

	l1.SetRune(0, 0, 'A', text.Style{})
	// l2 transparent
	l3.SetRune(0, 0, 'C', text.Style{})

	Composite(dst, []*Buffer{l1, l2, l3})
	// l3 is topmost and opaque, so it wins.
	if dst.Cell(0, 0).Rune != 'C' {
		t.Fatalf("got %c, want C", dst.Cell(0, 0).Rune)
	}
}

func TestCompositeMiddleLayerWins(t *testing.T) {
	dst := NewBuffer(1, 1)
	l1 := NewBuffer(1, 1)
	l2 := NewBuffer(1, 1)
	l3 := NewBuffer(1, 1) // transparent

	l1.SetRune(0, 0, 'A', text.Style{})
	l2.SetRune(0, 0, 'B', text.Style{})
	// l3 transparent

	Composite(dst, []*Buffer{l1, l2, l3})
	// l3 transparent, l2 opaque -> l2 wins.
	if dst.Cell(0, 0).Rune != 'B' {
		t.Fatalf("got %c, want B", dst.Cell(0, 0).Rune)
	}
}

func TestCompositeTwoStylePreserved(t *testing.T) {
	dst := NewBuffer(1, 1)
	base := NewBuffer(1, 1)
	overlay := NewBuffer(1, 1)

	s := text.Style{}.WithFg(text.Red()).Bold()
	overlay.SetRune(0, 0, 'R', s)

	CompositeTwo(dst, base, overlay)
	c := dst.Cell(0, 0)
	if !c.Style.Equal(s) {
		t.Fatal("CompositeTwo should preserve style from overlay")
	}
}
