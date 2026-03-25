package text

import "testing"

func TestRuneWidth_ASCII(t *testing.T) {
	for r := rune(0x20); r <= 0x7E; r++ {
		if w := RuneWidth(r); w != 1 {
			t.Errorf("RuneWidth(%q) = %d, want 1", r, w)
		}
	}
}

func TestRuneWidth_Control(t *testing.T) {
	controls := []rune{0x00, 0x01, 0x0A, 0x1F, 0x7F}
	for _, r := range controls {
		if w := RuneWidth(r); w != 0 {
			t.Errorf("RuneWidth(%U) = %d, want 0", r, w)
		}
	}
}

func TestRuneWidth_CJK(t *testing.T) {
	wideChars := []rune{
		0x4E2D, // CJK Unified (Chinese character)
		0x3042, // Hiragana 'a'
		0xAC00, // Hangul Syllable 'ga'
		0xFF01, // Fullwidth exclamation mark
		0x1100, // Hangul Jamo
		0xF900, // CJK Compat Ideograph
		0xFE30, // CJK Compat Form
		0xFFE0, // Fullwidth cent sign
	}
	for _, r := range wideChars {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%U) = %d, want 2", r, w)
		}
	}
}

func TestStringWidth_Mixed(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"hello", 5},
		{"", 0},
		{"abc", 3},
		{"\x00\x01", 0},                  // control chars
		{string([]rune{0x4E2D}), 2},      // single CJK
		{"a" + string([]rune{0x4E2D}), 3}, // ASCII + CJK
		{string([]rune{0x4E2D, 0x6587}), 4}, // two CJK chars
	}
	for _, tt := range tests {
		if got := StringWidth(tt.s); got != tt.want {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestStyledLineWidth(t *testing.T) {
	line := StyledLine{
		{Text: "hello", Style: Style{}},
		{Text: string([]rune{0x4E16, 0x754C})}, // two CJK chars = 4
	}
	if got := StyledLineWidth(line); got != 9 {
		t.Errorf("StyledLineWidth = %d, want 9", got)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s        string
		maxWidth int
		want     string
	}{
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		// CJK: each char is width 2
		{string([]rune{0x4E2D, 0x6587}), 4, string([]rune{0x4E2D, 0x6587})},
		{string([]rune{0x4E2D, 0x6587}), 3, string([]rune{0x4E2D})}, // can't fit second wide char
		{string([]rune{0x4E2D, 0x6587}), 1, ""},                      // can't fit even one wide char
		// Mixed
		{"a" + string([]rune{0x4E2D}) + "b", 3, "a" + string([]rune{0x4E2D})},
	}
	for _, tt := range tests {
		if got := Truncate(tt.s, tt.maxWidth); got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxWidth, got, tt.want)
		}
	}
}

func TestStyledTruncate(t *testing.T) {
	bold := Style{}.Bold()
	line := StyledLine{
		{Text: "hello", Style: Style{}},
		{Text: " world", Style: bold},
	}

	got := StyledTruncate(line, 8)
	if len(got) != 2 {
		t.Fatalf("StyledTruncate: got %d spans, want 2", len(got))
	}
	if got[0].Text != "hello" {
		t.Errorf("span 0 text = %q, want %q", got[0].Text, "hello")
	}
	if got[1].Text != " wo" {
		t.Errorf("span 1 text = %q, want %q", got[1].Text, " wo")
	}
	if got[1].Style != bold {
		t.Error("span 1 style not preserved")
	}
}

func TestStyledTruncate_Zero(t *testing.T) {
	line := StyledLine{{Text: "hello"}}
	got := StyledTruncate(line, 0)
	if got != nil {
		t.Errorf("StyledTruncate(_, 0) = %v, want nil", got)
	}
}
