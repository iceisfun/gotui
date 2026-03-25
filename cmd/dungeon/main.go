package main

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/iceisfun/gotui/pkg/input"
	"github.com/iceisfun/gotui/pkg/layout"
	"github.com/iceisfun/gotui/pkg/render"
	"github.com/iceisfun/gotui/pkg/term"
	"github.com/iceisfun/gotui/pkg/text"
)

const (
	mapW = 200
	mapH = 200

	sightRadius  = 8
	edgeMargin   = 5
	wallFillPct  = 45
	smoothPasses = 5
)

// Tile types.
const (
	tileWall  = '#'
	tileFloor = '.'
	tileWater = '~'
	tileDoor  = '+'
	tileStair = '>'
)

// dungeon holds the generated map state.
type dungeon struct {
	tiles   [mapW * mapH]byte
	visited [mapW * mapH]bool

	playerX, playerY int
	moves            int

	surface *render.Surface

	// Cached viewport size for NearEdge checks (updated each render).
	viewW, viewH int
}

func newDungeon() *dungeon {
	d := &dungeon{}
	d.generate()
	d.surface = render.NewSurface(d.cellAt)
	d.surface.CenterOn(d.playerX, d.playerY)
	d.updateVisibility()
	return d
}

// idx converts 2D map coordinates to a flat index.
func idx(x, y int) int { return y*mapW + x }

// inBounds reports whether (x,y) is inside the map.
func inBounds(x, y int) bool {
	return x >= 0 && x < mapW && y >= 0 && y < mapH
}

// generate creates the dungeon using cellular automata.
func (d *dungeon) generate() {
	// Step 1: random fill.
	for i := range d.tiles {
		if rand.IntN(100) < wallFillPct {
			d.tiles[i] = tileWall
		} else {
			d.tiles[i] = tileFloor
		}
	}
	// Ensure borders are walls.
	for x := 0; x < mapW; x++ {
		d.tiles[idx(x, 0)] = tileWall
		d.tiles[idx(x, mapH-1)] = tileWall
	}
	for y := 0; y < mapH; y++ {
		d.tiles[idx(0, y)] = tileWall
		d.tiles[idx(mapW-1, y)] = tileWall
	}

	// Step 2: cellular automata smoothing.
	var buf [mapW * mapH]byte
	for pass := 0; pass < smoothPasses; pass++ {
		copy(buf[:], d.tiles[:])
		for y := 1; y < mapH-1; y++ {
			for x := 1; x < mapW-1; x++ {
				walls := 0
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						if buf[idx(x+dx, y+dy)] == tileWall {
							walls++
						}
					}
				}
				if walls >= 5 {
					d.tiles[idx(x, y)] = tileWall
				} else {
					d.tiles[idx(x, y)] = tileFloor
				}
			}
		}
	}

	// Step 3: sprinkle water — flood-fill a few small areas.
	for i := 0; i < 8; i++ {
		sx := rand.IntN(mapW-4) + 2
		sy := rand.IntN(mapH-4) + 2
		if d.tiles[idx(sx, sy)] == tileFloor {
			d.floodWater(sx, sy, 20+rand.IntN(30))
		}
	}

	// Step 4: place some doors on floor tiles adjacent to walls.
	doorsPlaced := 0
	for doorsPlaced < 30 {
		x := rand.IntN(mapW-2) + 1
		y := rand.IntN(mapH-2) + 1
		if d.tiles[idx(x, y)] != tileFloor {
			continue
		}
		// Must have walls on two opposite sides and floor on the other two (or vice versa).
		nWall := 0
		for _, nb := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			if d.tiles[idx(x+nb[0], y+nb[1])] == tileWall {
				nWall++
			}
		}
		if nWall == 2 {
			d.tiles[idx(x, y)] = tileDoor
			doorsPlaced++
		}
	}

	// Step 5: place stairs on a random floor tile.
	for {
		x := rand.IntN(mapW-2) + 1
		y := rand.IntN(mapH-2) + 1
		if d.tiles[idx(x, y)] == tileFloor {
			d.tiles[idx(x, y)] = tileStair
			break
		}
	}

	// Step 6: place player on a random floor tile.
	for {
		x := rand.IntN(mapW-2) + 1
		y := rand.IntN(mapH-2) + 1
		if d.tiles[idx(x, y)] == tileFloor {
			d.playerX = x
			d.playerY = y
			break
		}
	}
}

// floodWater does a limited BFS flood-fill of water tiles.
func (d *dungeon) floodWater(sx, sy, maxTiles int) {
	type pos struct{ x, y int }
	queue := []pos{{sx, sy}}
	seen := make(map[pos]bool)
	seen[pos{sx, sy}] = true
	filled := 0

	for len(queue) > 0 && filled < maxTiles {
		p := queue[0]
		queue = queue[1:]
		if d.tiles[idx(p.x, p.y)] != tileFloor {
			continue
		}
		d.tiles[idx(p.x, p.y)] = tileWater
		filled++

		for _, nb := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			np := pos{p.x + nb[0], p.y + nb[1]}
			if inBounds(np.x, np.y) && !seen[np] {
				seen[np] = true
				queue = append(queue, np)
			}
		}
	}
}

// updateVisibility marks tiles within sight radius as visited.
func (d *dungeon) updateVisibility() {
	for dy := -sightRadius; dy <= sightRadius; dy++ {
		for dx := -sightRadius; dx <= sightRadius; dx++ {
			wx := d.playerX + dx
			wy := d.playerY + dy
			if !inBounds(wx, wy) {
				continue
			}
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(sightRadius) {
				d.visited[idx(wx, wy)] = true
			}
		}
	}
}

// isVisible reports whether a world tile is currently visible to the player.
func (d *dungeon) isVisible(wx, wy int) bool {
	dx := wx - d.playerX
	dy := wy - d.playerY
	return dx*dx+dy*dy <= sightRadius*sightRadius
}

// tileStyle returns the style for a tile at full brightness.
func tileStyle(tile byte) (rune, text.Style) {
	switch tile {
	case tileWall:
		return '#', text.Style{
			Fg: text.Color256(242),
			Bg: text.Color256(237),
		}
	case tileFloor:
		return '.', text.Style{
			Fg: text.Color256(242),
			Bg: text.Color256(233),
		}
	case tileWater:
		return '~', text.Style{
			Fg: text.Color256(39),
			Bg: text.Color256(17),
		}
	case tileDoor:
		return '+', text.Style{
			Fg: text.Yellow(),
			Bg: text.Color256(233),
		}
	case tileStair:
		return '>', text.Style{
			Fg: text.BrightMagenta(),
			Bg: text.Color256(233),
		}
	default:
		return ' ', text.Style{}
	}
}

// dimStyle returns a dimmed version of a style for fog of war.
func dimStyle(s text.Style) text.Style {
	return s.Dim()
}

// cellAt is the Surface callback: returns the cell for a given world position.
func (d *dungeon) cellAt(wx, wy int) render.Cell {
	// Player position.
	if wx == d.playerX && wy == d.playerY {
		return render.Cell{
			Rune:  '@',
			Width: 1,
			Style: text.Style{Fg: text.BrightWhite()}.Bold(),
		}
	}

	if !inBounds(wx, wy) {
		return render.Cell{Transparent: true}
	}

	tile := d.tiles[idx(wx, wy)]
	visible := d.isVisible(wx, wy)
	wasVisited := d.visited[idx(wx, wy)]

	if visible {
		r, s := tileStyle(tile)
		return render.Cell{Rune: r, Width: 1, Style: s}
	}
	if wasVisited {
		r, s := tileStyle(tile)
		return render.Cell{Rune: r, Width: 1, Style: dimStyle(s)}
	}

	// Never seen: black/transparent.
	return render.Cell{Rune: ' ', Width: 1, Style: text.Style{}}
}

// tryMove attempts to move the player by (dx, dy).
func (d *dungeon) tryMove(dx, dy int) {
	nx, ny := d.playerX+dx, d.playerY+dy
	if !inBounds(nx, ny) {
		return
	}
	tile := d.tiles[idx(nx, ny)]
	if tile == tileWall {
		return
	}
	d.playerX = nx
	d.playerY = ny
	d.moves++
	d.updateVisibility()

	if d.surface.NearEdge(d.playerX, d.playerY, edgeMargin, d.viewW, d.viewH) {
		d.surface.CenterOn(d.playerX, d.playerY)
	}
}

// tileNameAt returns a human-readable name for the tile under the player.
func (d *dungeon) tileNameAt() string {
	if !inBounds(d.playerX, d.playerY) {
		return "void"
	}
	switch d.tiles[idx(d.playerX, d.playerY)] {
	case tileFloor:
		return "floor"
	case tileWall:
		return "wall"
	case tileWater:
		return "water"
	case tileDoor:
		return "door"
	case tileStair:
		return "stairs"
	default:
		return "unknown"
	}
}

// dungeonView wraps the dungeon for rendering and input handling.
type dungeonView struct {
	d       *dungeon
	focused bool
}

func (dv *dungeonView) Render(v *render.View) {
	dv.d.viewW = v.Width()
	dv.d.viewH = v.Height()
	v.Clear()
	dv.d.surface.Render(v)
}

func (dv *dungeonView) HandleEvent(ev input.Event) bool {
	if ev.Type != input.EventKey {
		return false
	}
	k := ev.Key

	dx, dy := 0, 0

	switch {
	// Arrow keys
	case k.Code == input.KeyUp:
		dy = -1
	case k.Code == input.KeyDown:
		dy = 1
	case k.Code == input.KeyLeft:
		dx = -1
	case k.Code == input.KeyRight:
		dx = 1
	// WASD
	case k.Code == input.KeyRune && k.Rune == 'w':
		dy = -1
	case k.Code == input.KeyRune && k.Rune == 's':
		dy = 1
	case k.Code == input.KeyRune && k.Rune == 'a':
		dx = -1
	case k.Code == input.KeyRune && k.Rune == 'd':
		dx = 1
	// Vi keys
	case k.Code == input.KeyRune && k.Rune == 'k':
		dy = -1
	case k.Code == input.KeyRune && k.Rune == 'j':
		dy = 1
	case k.Code == input.KeyRune && k.Rune == 'h':
		dx = -1
	case k.Code == input.KeyRune && k.Rune == 'l':
		dx = 1
	default:
		return false
	}

	dv.d.tryMove(dx, dy)
	return true
}

func (dv *dungeonView) Focus()            { dv.focused = true }
func (dv *dungeonView) Blur()             { dv.focused = false }
func (dv *dungeonView) IsFocused() bool   { return dv.focused }

// statusBar shows player info along the bottom.
type statusBar struct {
	d *dungeon
}

func (sb *statusBar) Render(v *render.View) {
	w := v.Width()
	style := text.Style{Fg: text.BrightWhite(), Bg: text.Color256(236)}
	// Fill both rows with the background.
	for x := 0; x < w; x++ {
		v.SetRune(x, 0, ' ', style)
		v.SetRune(x, 1, ' ', style)
	}

	line1 := fmt.Sprintf(" Pos: (%d,%d)  Tile: %s  Moves: %d",
		sb.d.playerX, sb.d.playerY, sb.d.tileNameAt(), sb.d.moves)
	line2 := " wasd/hjkl/arrows to move, q to quit"

	v.WriteString(0, 0, line1, style)
	v.WriteString(0, 1, line2, style.Dim())
}

func main() {
	t, err := term.Open()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	d := newDungeon()
	dv := &dungeonView{d: d}
	sb := &statusBar{d: d}

	root := layout.NewVSplit(
		layout.SplitChild{Renderable: dv, Size: 1.0},
		layout.SplitChild{Renderable: sb, Size: 2},
	)

	app := layout.NewApp(t, root)
	app.SetFocus(dv)

	// Quit on 'q'.
	app.BindRune('q', 0, func() { app.Quit() })

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
