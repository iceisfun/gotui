# cmd/dungeon -- Dungeon Explorer

An Angband-style procedural cave exploration game built on the GoREPL
rendering engine. The player (@) explores a 200x200 tile map with fog of
war, navigating through caves, water, doors, and stairs.

## How to Run

```
go run ./cmd/dungeon
```

## Controls

| Key | Action |
|-----|--------|
| W / k / Up | Move north |
| S / j / Down | Move south |
| A / h / Left | Move west |
| D / l / Right | Move east |
| q | Quit |

## Map Generation

The dungeon is generated using cellular automata:

1. **Random fill**: Each tile is randomly assigned as wall or floor based on a
   45% wall density.
2. **Smoothing**: Five passes of cellular automata where a tile becomes wall if
   5 or more of its 8 neighbors are walls, producing natural cave formations.
3. **Water**: Small pools are flood-filled onto random floor areas.
4. **Doors**: Placed on floor tiles that have walls on exactly two sides.
5. **Stairs**: A single stair tile (>) placed on a random floor.

## Tile Types

| Symbol | Tile |
|--------|------|
| # | Wall (impassable) |
| . | Floor |
| ~ | Water (passable) |
| + | Door (passable) |
| > | Stairs |
| @ | Player |

## Visibility

The player has a circular line-of-sight with radius 8. Tiles within range are
fully lit; previously seen tiles appear dimmed (fog of war); unexplored tiles
are black. The camera auto-scrolls when the player approaches the viewport
edge.
