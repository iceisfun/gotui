package repl

import "github.com/iceisfun/gorepl/pkg/text"

// Executor runs user input and returns output. The host application provides
// the implementation — the REPL package has no dependency on any language.
type Executor interface {
	// Execute runs source and returns the output string and any error.
	Execute(source string) (output string, err error)

	// IsComplete reports whether source is a complete statement.
	// Returns false if the user needs to type more lines (e.g., unclosed block).
	IsComplete(source string) bool
}

// Completer provides tab-completion proposals.
type Completer interface {
	// Complete returns proposals for the given line at the cursor position.
	Complete(line string, pos int) []Completion
}

// Completion is a single completion proposal.
type Completion struct {
	Text    string         // Text to insert.
	Display text.StyledLine // How to render in the popup.
	Detail  string         // Optional type/detail info.
}

// Highlighter provides live syntax coloring of the input buffer.
type Highlighter interface {
	// Highlight returns a styled version of one line of source.
	Highlight(line string) text.StyledLine
}
