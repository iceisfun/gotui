package render

// Composite flattens a stack of layers into a destination buffer.
// Layers are ordered bottom-to-top. For each cell position, the topmost
// non-transparent cell wins. All buffers must have the same dimensions.
func Composite(dst *Buffer, layers []*Buffer) {
	if len(layers) == 0 {
		return
	}

	w, h := dst.Width(), dst.Height()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			// Walk top-to-bottom, take the first non-transparent cell.
			for i := len(layers) - 1; i >= 0; i-- {
				c := layers[i].cells[idx]
				if !c.Transparent {
					dst.cells[idx] = c
					break
				}
				if i == 0 {
					// All layers transparent at this position — use blank.
					dst.cells[idx] = BlankCell
				}
			}
		}
	}
}

// CompositeTwo is an optimized path for the common case of exactly two layers
// (base + one overlay).
func CompositeTwo(dst *Buffer, base, overlay *Buffer) {
	w, h := dst.Width(), dst.Height()
	for i := 0; i < w*h; i++ {
		c := overlay.cells[i]
		if c.Transparent {
			c = base.cells[i]
		}
		dst.cells[i] = c
	}
}
