package layout

import (
	"slices"

	"github.com/iceisfun/gorepl/pkg/render"
)

// overlayEntry is an overlay with its absolute screen position.
type overlayEntry struct {
	renderable render.Renderable
	bounds     render.Rect
	zOrder     int
}

// CollectOverlays walks a renderable tree and collects all OverlayRequests,
// translating their anchors to absolute screen coordinates using the given
// parent offset.
func CollectOverlays(r render.Renderable, parentBounds render.Rect) []overlayEntry {
	var entries []overlayEntry

	if o, ok := r.(render.Overlayable); ok {
		for _, req := range o.Overlays() {
			abs := req.Anchor.Translate(parentBounds.X, parentBounds.Y)
			entries = append(entries, overlayEntry{
				renderable: req.Renderable,
				bounds:     abs,
				zOrder:     req.ZOrder,
			})
		}
	}

	// Recurse into container children.
	if c, ok := r.(render.Container); ok {
		childBounds := c.ChildBounds()
		children := c.Children()
		for i, child := range children {
			if i < len(childBounds) {
				entries = append(entries, CollectOverlays(child, childBounds[i])...)
			}
		}
	}

	return entries
}

// SortOverlays orders overlay entries by z-order (lowest first).
func SortOverlays(entries []overlayEntry) {
	slices.SortStableFunc(entries, func(a, b overlayEntry) int {
		return a.zOrder - b.zOrder
	})
}
