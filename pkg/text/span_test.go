package text

import "testing"

func TestStyledLineLen(t *testing.T) {
	tests := []struct {
		name string
		line StyledLine
		want int
	}{
		{"empty", StyledLine{}, 0},
		{"single ascii", Plain("hello"), 5},
		{"multi-byte", Plain("\u00e9\u00e9"), 2}, // two é characters
		{"two spans", StyledLine{{Text: "ab"}, {Text: "cd"}}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.line.Len(); got != tt.want {
				t.Fatalf("Len() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStyledLinePlainText(t *testing.T) {
	tests := []struct {
		name string
		line StyledLine
		want string
	}{
		{"empty", StyledLine{}, ""},
		{"single span", Plain("hello"), "hello"},
		{"two spans", StyledLine{{Text: "he"}, {Text: "llo"}}, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.line.PlainText(); got != tt.want {
				t.Fatalf("PlainText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlainConstructor(t *testing.T) {
	line := Plain("test")
	if len(line) != 1 {
		t.Fatalf("expected 1 span, got %d", len(line))
	}
	if line[0].Text != "test" {
		t.Fatalf("text = %q", line[0].Text)
	}
	if !line[0].Style.Equal(Style{}) {
		t.Fatal("Plain should use default style")
	}
}

func TestStyledConstructor(t *testing.T) {
	s := Style{}.WithFg(Red()).Bold()
	line := Styled("test", s)
	if len(line) != 1 {
		t.Fatalf("expected 1 span, got %d", len(line))
	}
	if line[0].Text != "test" {
		t.Fatalf("text = %q", line[0].Text)
	}
	if !line[0].Style.Equal(s) {
		t.Fatal("style mismatch")
	}
}

func TestStyledLineAppend(t *testing.T) {
	line := Plain("hello")
	line = line.Append(" world", Style{}.Bold())
	if line.PlainText() != "hello world" {
		t.Fatalf("PlainText() = %q", line.PlainText())
	}
	if len(line) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(line))
	}
}

func TestStyledLineLenMultiByte(t *testing.T) {
	// CJK characters are multi-byte in UTF-8 but each is one rune.
	line := Plain("\u4e16\u754c") // 世界
	if line.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", line.Len())
	}
}
