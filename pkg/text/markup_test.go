package text

import "testing"

func TestParsePlainText(t *testing.T) {
	line := Parse("hello world")
	if len(line) != 1 {
		t.Fatalf("expected 1 span, got %d", len(line))
	}
	if line[0].Text != "hello world" {
		t.Fatalf("text = %q", line[0].Text)
	}
	if !line[0].Style.Equal(Style{}) {
		t.Fatal("expected default style")
	}
}

func TestParseEmpty(t *testing.T) {
	line := Parse("")
	if len(line) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(line))
	}
}

func TestParseSingleTag(t *testing.T) {
	line := Parse("[bold]text[/]")
	if line.PlainText() != "text" {
		t.Fatalf("PlainText() = %q, want %q", line.PlainText(), "text")
	}
	if line[0].Style.Attrs&AttrBold == 0 {
		t.Fatal("expected bold attribute")
	}
}

func TestParseNestedTags(t *testing.T) {
	line := Parse("[bold][red]text[/][/]")
	if line.PlainText() != "text" {
		t.Fatalf("PlainText() = %q", line.PlainText())
	}
	if line[0].Style.Attrs&AttrBold == 0 {
		t.Fatal("expected bold")
	}
	if !line[0].Style.Fg.Equal(Red()) {
		t.Fatal("expected red foreground")
	}
}

func TestParseNestedRestore(t *testing.T) {
	line := Parse("[bold]A[red]B[/]C[/]D")
	if line.PlainText() != "ABCD" {
		t.Fatalf("PlainText() = %q", line.PlainText())
	}
	// A: bold, default fg
	if line[0].Style.Attrs&AttrBold == 0 {
		t.Fatal("A should be bold")
	}
	if !line[0].Style.Fg.IsDefault() {
		t.Fatal("A should have default fg")
	}
	// B: bold + red
	if line[1].Style.Attrs&AttrBold == 0 {
		t.Fatal("B should be bold")
	}
	if !line[1].Style.Fg.Equal(Red()) {
		t.Fatal("B should be red")
	}
	// C: bold (popped red)
	if line[2].Style.Attrs&AttrBold == 0 {
		t.Fatal("C should be bold")
	}
	if !line[2].Style.Fg.IsDefault() {
		t.Fatal("C should have default fg after pop")
	}
	// D: default (popped bold)
	if line[3].Style.Attrs != 0 {
		t.Fatal("D should have no attributes")
	}
}

func TestParseNamedColors(t *testing.T) {
	tests := []struct {
		tag  string
		want Color
	}{
		{"red", Red()},
		{"green", Green()},
		{"blue", Blue()},
		{"yellow", Yellow()},
		{"cyan", Cyan()},
		{"magenta", Magenta()},
		{"white", White()},
		{"black", Black()},
		{"bright_red", BrightRed()},
		{"bright_blue", BrightBlue()},
	}
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			line := Parse("[" + tt.tag + "]x[/]")
			if !line[0].Style.Fg.Equal(tt.want) {
				t.Fatalf("expected %v fg", tt.want)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	line := Parse("[#ff8800]text[/]")
	want := RGB(0xff, 0x88, 0x00)
	if !line[0].Style.Fg.Equal(want) {
		t.Fatalf("Fg = %v, want RGB(255,136,0)", line[0].Style.Fg)
	}
}

func TestParseBackgroundNamed(t *testing.T) {
	line := Parse("[on blue]text[/]")
	if !line[0].Style.Bg.Equal(Blue()) {
		t.Fatal("expected blue background")
	}
	if !line[0].Style.Fg.IsDefault() {
		t.Fatal("fg should remain default")
	}
}

func TestParseBackgroundHex(t *testing.T) {
	line := Parse("[on #003366]text[/]")
	want := RGB(0x00, 0x33, 0x66)
	if !line[0].Style.Bg.Equal(want) {
		t.Fatalf("Bg mismatch")
	}
}

func TestParseCombined(t *testing.T) {
	line := Parse("[bold italic red on blue]text[/]")
	s := line[0].Style
	if s.Attrs&AttrBold == 0 {
		t.Fatal("expected bold")
	}
	if s.Attrs&AttrItalic == 0 {
		t.Fatal("expected italic")
	}
	if !s.Fg.Equal(Red()) {
		t.Fatal("expected red fg")
	}
	if !s.Bg.Equal(Blue()) {
		t.Fatal("expected blue bg")
	}
}

func TestParseUnclosedBracket(t *testing.T) {
	// Should not panic; unclosed bracket treated as literal text.
	line := Parse("hello [bold")
	if line.PlainText() != "hello [bold" {
		t.Fatalf("PlainText() = %q, want %q", line.PlainText(), "hello [bold")
	}
}

func TestParseMultipleSpans(t *testing.T) {
	line := Parse("plain [bold]bold[/] more")
	if line.PlainText() != "plain bold more" {
		t.Fatalf("PlainText() = %q", line.PlainText())
	}
	if len(line) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(line))
	}
	// first span: plain
	if line[0].Style.Attrs != 0 {
		t.Fatal("first span should be unstyled")
	}
	// second span: bold
	if line[1].Style.Attrs&AttrBold == 0 {
		t.Fatal("second span should be bold")
	}
	// third span: after pop, back to default
	if line[2].Style.Attrs != 0 {
		t.Fatal("third span should be unstyled")
	}
}

func TestParseExtraCloseTags(t *testing.T) {
	// More [/] than opens should reset to default, not panic.
	line := Parse("[bold]text[/][/][/]rest")
	if line.PlainText() != "textrest" {
		t.Fatalf("PlainText() = %q", line.PlainText())
	}
}

func TestParseAllAttributes(t *testing.T) {
	line := Parse("[bold dim italic underline blink reverse strikethrough]x[/]")
	s := line[0].Style
	want := AttrBold | AttrDim | AttrItalic | AttrUnderline | AttrBlink | AttrReverse | AttrStrikethrough
	if s.Attrs != want {
		t.Fatalf("Attrs = %v, want %v", s.Attrs, want)
	}
}
