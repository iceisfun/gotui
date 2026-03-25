package render

// Surface is an infinite scrollable virtual canvas. The user provides a CellAt
// callback that generates content for any world coordinate on demand. A camera
// position controls which portion of the world is visible in the viewport.
type Surface struct {
	// CellAt returns the cell to display at the given world coordinates.
	// A zero-value Cell (Transparent == false, Rune == 0) will be rendered as
	// a blank space. Return a TransparentCell to leave the position empty.
	CellAt func(worldX, worldY int) Cell

	cameraX, cameraY int // world coordinate of the top-left corner of the viewport
}

// NewSurface creates a Surface with the given cell provider.
func NewSurface(cellAt func(worldX, worldY int) Cell) *Surface {
	return &Surface{CellAt: cellAt}
}

// CenterOn sets the camera so that the given world coordinate will appear at
// the center of the viewport. Because the viewport size is not known until
// Render is called, CenterOn stores the desired center and Render adjusts.
// We store a "desired center" and convert to top-left in Render.
func (s *Surface) CenterOn(x, y int) {
	// Store as a sentinel; Render will subtract half the viewport size.
	// We use negative-max as a flag... actually, simpler to just store the
	// center and convert in Render.
	s.cameraX = x
	s.cameraY = y
}

// Camera returns the current camera position (world coordinate of viewport center).
func (s *Surface) Camera() (int, int) {
	return s.cameraX, s.cameraY
}

// NearEdge reports whether the world position (x, y) is within margin cells of
// the viewport edge. The viewport size is given explicitly so that callers can
// check before rendering.
func (s *Surface) NearEdge(x, y, margin, viewW, viewH int) bool {
	// Top-left of viewport in world coords.
	tlx := s.cameraX - viewW/2
	tly := s.cameraY - viewH/2

	localX := x - tlx
	localY := y - tly

	return localX < margin || localX >= viewW-margin ||
		localY < margin || localY >= viewH-margin
}

// WorldToScreen converts a world coordinate to a screen (view-relative)
// coordinate given the current camera and viewport dimensions.
func (s *Surface) WorldToScreen(wx, wy, viewW, viewH int) (sx, sy int) {
	tlx := s.cameraX - viewW/2
	tly := s.cameraY - viewH/2
	return wx - tlx, wy - tly
}

// ScreenToWorld converts a screen (view-relative) coordinate to a world
// coordinate given the current camera and viewport dimensions.
func (s *Surface) ScreenToWorld(sx, sy, viewW, viewH int) (wx, wy int) {
	tlx := s.cameraX - viewW/2
	tly := s.cameraY - viewH/2
	return sx + tlx, sy + tly
}

// Render draws the visible portion of the world into the provided View. It
// implements the Renderable interface.
func (s *Surface) Render(v *View) {
	w, h := v.Width(), v.Height()
	tlx := s.cameraX - w/2
	tly := s.cameraY - h/2

	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			c := s.CellAt(tlx+sx, tly+sy)
			if c.Transparent {
				continue
			}
			v.SetCell(sx, sy, c)
		}
	}
}
