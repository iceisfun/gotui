package text

import (
	"reflect"
	"testing"
)

func TestWrap_SimpleWords(t *testing.T) {
	got := Wrap("hello world foo", 11)
	want := []string{"hello world", "foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap = %q, want %q", got, want)
	}
}

func TestWrap_ExactFit(t *testing.T) {
	got := Wrap("hello world", 11)
	want := []string{"hello world"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap = %q, want %q", got, want)
	}
}

func TestWrap_LongWordBreaking(t *testing.T) {
	got := Wrap("abcdefghij", 4)
	want := []string{"abcd", "efgh", "ij"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap = %q, want %q", got, want)
	}
}

func TestWrap_PreservesNewlines(t *testing.T) {
	got := Wrap("hello\nworld", 20)
	want := []string{"hello", "world"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap = %q, want %q", got, want)
	}
}

func TestWrap_CJK(t *testing.T) {
	// Each CJK char is width 2. Four chars = width 8.
	s := string([]rune{0x4E2D, 0x6587, 0x6D4B, 0x8BD5})
	got := Wrap(s, 5) // fits 2 chars (width 4), then 2 chars (width 4)
	want := []string{
		string([]rune{0x4E2D, 0x6587}),
		string([]rune{0x6D4B, 0x8BD5}),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap CJK = %q, want %q", got, want)
	}
}

func TestWrap_Empty(t *testing.T) {
	got := Wrap("", 10)
	want := []string{""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap empty = %q, want %q", got, want)
	}
}

func TestWrap_ZeroWidth(t *testing.T) {
	got := Wrap("hello", 0)
	if got != nil {
		t.Errorf("Wrap(_, 0) = %v, want nil", got)
	}
}

func TestWrapStyled_PreservesStyles(t *testing.T) {
	bold := Style{}.Bold()
	line := StyledLine{
		{Text: "hello ", Style: Style{}},
		{Text: "world", Style: bold},
	}

	got := WrapStyled(line, 7)
	if len(got) != 2 {
		t.Fatalf("WrapStyled: got %d lines, want 2", len(got))
	}

	// First line: "hello" (plain) + space handled via word boundary
	// Second line: "world" (bold)
	line1 := got[0].PlainText()
	line2 := got[1].PlainText()
	if line1 != "hello" {
		t.Errorf("line 0 = %q, want %q", line1, "hello")
	}
	if line2 != "world" {
		t.Errorf("line 1 = %q, want %q", line2, "world")
	}

	// Check that "world" on line 2 has the bold style.
	if len(got[1]) == 0 {
		t.Fatal("line 1 has no spans")
	}
	if got[1][0].Style != bold {
		t.Error("line 1 style not bold")
	}
}

func TestWrapStyled_Empty(t *testing.T) {
	got := WrapStyled(StyledLine{}, 10)
	if len(got) != 1 {
		t.Fatalf("WrapStyled empty: got %d lines, want 1", len(got))
	}
}

func TestWrapStyled_WithNewlines(t *testing.T) {
	line := StyledLine{
		{Text: "hello\nworld", Style: Style{}},
	}
	got := WrapStyled(line, 20)
	if len(got) != 2 {
		t.Fatalf("WrapStyled newlines: got %d lines, want 2", len(got))
	}
	if got[0].PlainText() != "hello" {
		t.Errorf("line 0 = %q, want %q", got[0].PlainText(), "hello")
	}
	if got[1].PlainText() != "world" {
		t.Errorf("line 1 = %q, want %q", got[1].PlainText(), "world")
	}
}
