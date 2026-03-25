package repl

import "strings"

// History stores command history with navigation and reverse search.
type History struct {
	entries []string
	pos     int    // Current position during navigation (-1 = not navigating).
	saved   string // The in-progress input saved when navigation starts.
	maxSize int
}

// NewHistory creates a history with the given max entries.
func NewHistory(maxSize int) *History {
	return &History{
		maxSize: maxSize,
		pos:     -1,
	}
}

// Add appends an entry to history. Empty and duplicate-of-last entries are skipped.
func (h *History) Add(entry string) {
	entry = strings.TrimRight(entry, "\n")
	if entry == "" {
		return
	}
	// Skip consecutive duplicates.
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		h.ResetNav()
		return
	}
	h.entries = append(h.entries, entry)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
	h.ResetNav()
}

// Prev moves to the previous (older) entry. Returns the entry and true,
// or empty string and false if at the beginning.
// currentInput is saved on the first call so it can be restored.
func (h *History) Prev(currentInput string) (string, bool) {
	if len(h.entries) == 0 {
		return "", false
	}

	if h.pos == -1 {
		// Start navigating from the end.
		h.saved = currentInput
		h.pos = len(h.entries) - 1
	} else if h.pos > 0 {
		h.pos--
	} else {
		return h.entries[0], false
	}

	return h.entries[h.pos], true
}

// Next moves to the next (newer) entry. Returns the entry and true,
// or the saved input and false if past the end.
func (h *History) Next(currentInput string) (string, bool) {
	if h.pos == -1 {
		return currentInput, false
	}

	h.pos++
	if h.pos >= len(h.entries) {
		// Past end — restore saved input.
		result := h.saved
		h.ResetNav()
		return result, true
	}

	return h.entries[h.pos], true
}

// ResetNav resets navigation state.
func (h *History) ResetNav() {
	h.pos = -1
	h.saved = ""
}

// Search finds the most recent entry containing substr, searching backwards
// from the current position (or end if not navigating).
func (h *History) Search(substr string) (string, int, bool) {
	if substr == "" || len(h.entries) == 0 {
		return "", -1, false
	}

	start := len(h.entries) - 1
	if h.pos >= 0 && h.pos < start {
		start = h.pos - 1
	}

	for i := start; i >= 0; i-- {
		if strings.Contains(h.entries[i], substr) {
			h.pos = i
			return h.entries[i], i, true
		}
	}
	return "", -1, false
}

// Len returns the number of entries.
func (h *History) Len() int {
	return len(h.entries)
}
