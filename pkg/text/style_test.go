package text

import "testing"

func TestStyleWithFg(t *testing.T) {
	s := Style{}.WithFg(Red())
	if !s.Fg.Equal(Red()) {
		t.Fatal("WithFg did not set foreground")
	}
	if !s.Bg.IsDefault() {
		t.Fatal("WithFg changed background")
	}
}

func TestStyleWithBg(t *testing.T) {
	s := Style{}.WithBg(Blue())
	if !s.Bg.Equal(Blue()) {
		t.Fatal("WithBg did not set background")
	}
	if !s.Fg.IsDefault() {
		t.Fatal("WithBg changed foreground")
	}
}

func TestStyleWithAttr(t *testing.T) {
	s := Style{}.WithAttr(AttrBold)
	if s.Attrs&AttrBold == 0 {
		t.Fatal("WithAttr did not set bold")
	}
}

func TestStyleWithoutAttr(t *testing.T) {
	s := Style{}.WithAttr(AttrBold | AttrItalic).WithoutAttr(AttrBold)
	if s.Attrs&AttrBold != 0 {
		t.Fatal("WithoutAttr did not clear bold")
	}
	if s.Attrs&AttrItalic == 0 {
		t.Fatal("WithoutAttr cleared italic")
	}
}

func TestStyleBuilderChaining(t *testing.T) {
	s := Style{}.WithFg(Red()).WithBg(Blue()).Bold().Italic().Underline()
	if !s.Fg.Equal(Red()) {
		t.Fatal("Fg wrong")
	}
	if !s.Bg.Equal(Blue()) {
		t.Fatal("Bg wrong")
	}
	want := AttrBold | AttrItalic | AttrUnderline
	if s.Attrs != want {
		t.Fatalf("Attrs = %v, want %v", s.Attrs, want)
	}
}

func TestStyleShortcutMethods(t *testing.T) {
	tests := []struct {
		name string
		fn   func(Style) Style
		attr Attr
	}{
		{"Bold", Style.Bold, AttrBold},
		{"Dim", Style.Dim, AttrDim},
		{"Italic", Style.Italic, AttrItalic},
		{"Underline", Style.Underline, AttrUnderline},
		{"Reverse", Style.Reverse, AttrReverse},
		{"Strikethrough", Style.Strikethrough, AttrStrikethrough},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fn(Style{})
			if s.Attrs&tt.attr == 0 {
				t.Fatalf("%s() did not set the attribute", tt.name)
			}
		})
	}
}

func TestStyleEqual(t *testing.T) {
	a := Style{}.WithFg(Red()).Bold()
	b := Style{}.WithFg(Red()).Bold()
	c := Style{}.WithFg(Blue()).Bold()
	d := Style{}.WithFg(Red()).Italic()

	if !a.Equal(b) {
		t.Fatal("identical styles should be equal")
	}
	if a.Equal(c) {
		t.Fatal("different Fg should not be equal")
	}
	if a.Equal(d) {
		t.Fatal("different attrs should not be equal")
	}
}

func TestAttrBitmaskComposition(t *testing.T) {
	s := Style{}.WithAttr(AttrBold).WithAttr(AttrItalic).WithAttr(AttrReverse)
	if s.Attrs != AttrBold|AttrItalic|AttrReverse {
		t.Fatalf("attrs = %v, want %v", s.Attrs, AttrBold|AttrItalic|AttrReverse)
	}
	// Adding the same attr again should be idempotent.
	s = s.WithAttr(AttrBold)
	if s.Attrs != AttrBold|AttrItalic|AttrReverse {
		t.Fatal("re-adding attr changed the mask")
	}
}

func TestStyleImmutability(t *testing.T) {
	a := Style{}.WithFg(Red())
	b := a.Bold()
	if a.Attrs != 0 {
		t.Fatal("original style was mutated by Bold()")
	}
	if b.Attrs != AttrBold {
		t.Fatal("new style missing Bold")
	}
}
