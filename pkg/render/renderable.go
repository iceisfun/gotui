package render

import (
	"time"

	"github.com/iceisfun/gotui/pkg/input"
)

// Renderable is the minimum interface for anything that can draw itself.
type Renderable interface {
	Render(v *View)
}

// Measurable reports preferred, minimum, and maximum size.
// Discovered via type assertion.
type Measurable interface {
	Renderable
	MinSize() (w, h int)
	MaxSize() (w, h int)       // 0 = unlimited
	PreferredSize() (w, h int) // 0 = no preference
}

// Interactive handles input events.
// HandleEvent returns true if the event was consumed.
// Unconsumed events bubble up through the container tree.
type Interactive interface {
	Renderable
	HandleEvent(ev input.Event) bool
}

// Focusable can receive and lose keyboard focus.
// Keyboard events are dispatched to the focused component first.
type Focusable interface {
	Interactive
	Focus()
	Blur()
	IsFocused() bool
}

// Scrollable supports viewport scrolling over a larger content area.
type Scrollable interface {
	Renderable
	ContentSize() (w, h int)
	ScrollOffset() (x, y int)
	SetScrollOffset(x, y int)
}

// Cursorable controls the terminal cursor position and visibility.
// The layout system reads this to position the hardware cursor after rendering.
type Cursorable interface {
	Renderable
	CursorPosition() (x, y int, visible bool)
}

// Overlayable can produce floating renderables (completion popups, tooltips).
// The layout system collects these and renders them above the normal tree.
type Overlayable interface {
	Renderable
	Overlays() []OverlayRequest
}

// OverlayRequest describes a floating renderable anchored to its owner.
type OverlayRequest struct {
	Renderable Renderable
	Anchor     Rect // Position relative to the owning renderable's coordinate space.
	ZOrder     int  // Higher values render on top.
}

// Updatable receives periodic update ticks (for animation, polling, etc.).
type Updatable interface {
	Renderable
	Update(dt time.Duration)
}

// KeyBinding is a global keyboard accelerator.
// Registered on the App, these are checked before the focused renderable.
type KeyBinding struct {
	Key    input.KeyCode
	Mod    input.ModMask
	Rune   rune // For KeyRune bindings. Zero means match any rune with the given mods.
	Action func()
}

// MatchEvent reports whether this binding matches the given key event.
func (kb KeyBinding) MatchEvent(ev input.KeyEvent) bool {
	if ev.Mod != kb.Mod {
		return false
	}
	if kb.Key == input.KeyRune {
		return ev.Code == input.KeyRune && ev.Rune == kb.Rune
	}
	return ev.Code == kb.Key
}

// Container arranges child renderables within a region.
// Containers participate in the event bubble chain: if a child's HandleEvent
// returns false, the container gets a chance to handle the event.
type Container interface {
	Renderable
	Children() []Renderable
	Layout(bounds Rect)
	// ChildBounds returns the bounds assigned to each child after Layout.
	ChildBounds() []Rect
}

// MouseTarget is used by the layout system to hit-test mouse events.
// The App walks the container tree, finds the deepest renderable whose
// bounds contain the mouse coordinates, and dispatches to it. If that
// renderable does not consume the event, it bubbles up through parents.
// Focusable renderables receive focus on mouse click.
type MouseTarget interface {
	Interactive
	// HitTest reports whether this renderable should receive a mouse event
	// at the given coordinates (relative to its own origin).
	HitTest(x, y int) bool
}
