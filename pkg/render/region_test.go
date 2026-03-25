package render

import "testing"

func TestRectContains(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 30, Height: 40}
	tests := []struct {
		name string
		px   int
		py   int
		want bool
	}{
		{"inside", 15, 30, true},
		{"top-left corner", 10, 20, true},
		{"right edge excluded", 40, 30, false},
		{"bottom edge excluded", 15, 60, false},
		{"left of rect", 9, 30, false},
		{"above rect", 15, 19, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Contains(tt.px, tt.py); got != tt.want {
				t.Fatalf("Contains(%d,%d) = %v, want %v", tt.px, tt.py, got, tt.want)
			}
		})
	}
}

func TestRectIntersectOverlapping(t *testing.T) {
	a := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	b := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	got := a.Intersect(b)
	want := Rect{X: 5, Y: 5, Width: 5, Height: 5}
	if got != want {
		t.Fatalf("Intersect = %+v, want %+v", got, want)
	}
}

func TestRectIntersectNonOverlapping(t *testing.T) {
	a := Rect{X: 0, Y: 0, Width: 5, Height: 5}
	b := Rect{X: 10, Y: 10, Width: 5, Height: 5}
	got := a.Intersect(b)
	if !got.IsEmpty() {
		t.Fatalf("non-overlapping Intersect should be empty, got %+v", got)
	}
}

func TestRectIntersectAdjacent(t *testing.T) {
	a := Rect{X: 0, Y: 0, Width: 5, Height: 5}
	b := Rect{X: 5, Y: 0, Width: 5, Height: 5}
	got := a.Intersect(b)
	if !got.IsEmpty() {
		t.Fatalf("adjacent Intersect should be empty, got %+v", got)
	}
}

func TestRectIntersectContained(t *testing.T) {
	outer := Rect{X: 0, Y: 0, Width: 20, Height: 20}
	inner := Rect{X: 5, Y: 5, Width: 5, Height: 5}
	got := outer.Intersect(inner)
	if got != inner {
		t.Fatalf("Intersect = %+v, want %+v", got, inner)
	}
}

func TestRectIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"zero", Rect{}, true},
		{"zero width", Rect{Width: 0, Height: 5}, true},
		{"zero height", Rect{Width: 5, Height: 0}, true},
		{"negative width", Rect{Width: -1, Height: 5}, true},
		{"valid", Rect{Width: 1, Height: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsEmpty(); got != tt.want {
				t.Fatalf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRectTranslate(t *testing.T) {
	r := Rect{X: 5, Y: 10, Width: 20, Height: 30}
	got := r.Translate(3, -2)
	want := Rect{X: 8, Y: 8, Width: 20, Height: 30}
	if got != want {
		t.Fatalf("Translate = %+v, want %+v", got, want)
	}
}

func TestRectTranslatePreservesDimensions(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	got := r.Translate(100, 200)
	if got.Width != r.Width || got.Height != r.Height {
		t.Fatal("Translate should preserve width and height")
	}
}
