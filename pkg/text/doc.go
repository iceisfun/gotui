// Package text provides terminal text styling, color management, and markup parsing.
//
// Colors support 16-color, 256-color, and 24-bit truecolor modes via the [Color] type.
// The [Style] type combines foreground color, background color, and text attributes
// (bold, italic, underline, dim, reverse, strikethrough, blink).
//
// [StyledSpan] and [StyledLine] represent rich text as sequences of styled runs,
// suitable for rendering into terminal buffers.
//
// The [Parse] function converts bracket-tag markup like [bold red]text[/] into
// [StyledLine] values. Width measurement ([RuneWidth], [StringWidth]) and text
// wrapping ([Wrap], [WrapStyled]) are CJK-aware, handling double-width characters.
package text
