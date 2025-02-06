package pather

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather/astar"
)

type PathFinder struct {
	gr     *game.MemoryReader
	data   *game.Data
	hid    *game.HID
	cfg    *config.CharacterCfg
	logger *slog.Logger
}

func NewPathFinder(gr *game.MemoryReader, data *game.Data, hid *game.HID, cfg *config.CharacterCfg, logger *slog.Logger) *PathFinder {
	return &PathFinder{
		gr:     gr,
		data:   data,
		hid:    hid,
		cfg:    cfg,
		logger: logger,
	}
}

func (pf *PathFinder) GetPath(to data.Position) (Path, int, bool) {
	// First try direct path
	if path, distance, found := pf.GetPathFrom(pf.data.PlayerUnit.Position, to); found {
		return path, distance, true
	}

	// If direct path fails, try to find nearby walkable position
	if walkableTo, found := pf.findNearbyWalkablePosition(to); found {
		return pf.GetPathFrom(pf.data.PlayerUnit.Position, walkableTo)
	}

	return nil, 0, false
}

func (pf *PathFinder) GetPathFrom(from, to data.Position) (Path, int, bool) {
	a := pf.data.AreaData

	// We don't want to modify the original grid
	grid := a.Grid.Copy()

	// Special handling for Arcane Sanctuary (to allow pathing with platforms)
	if pf.data.PlayerUnit.Area == area.ArcaneSanctuary && pf.data.CanTeleport() {
		// Make all non-walkable tiles into low priority tiles for teleport pathing
		for y := 0; y < len(grid.CollisionGrid); y++ {
			for x := 0; x < len(grid.CollisionGrid[y]); x++ {
				if grid.CollisionGrid[y][x] == game.CollisionTypeNonWalkable {
					grid.CollisionGrid[y][x] = game.CollisionTypeLowPriority
				}
			}
		}
	}
	// Lut Gholein map is a bit bugged, we should close this fake path to avoid pathing issues
	if a.Area == area.LutGholein {
		a.CollisionGrid[13][210] = game.CollisionTypeNonWalkable
	}

	if !a.IsInside(to) {
		expandedGrid, err := pf.mergeGrids(to)
		if err != nil {
			return nil, 0, false
		}
		grid = expandedGrid
	}

	from = grid.RelativePosition(from)
	to = grid.RelativePosition(to)

	// Add objects to the collision grid as obstacles
	for _, o := range pf.data.AreaData.Objects {
		if !grid.IsWalkable(o.Position) {
			continue
		}
		relativePos := grid.RelativePosition(o.Position)
		grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeObject
		for i := -2; i <= 2; i++ {
			for j := -2; j <= 2; j++ {
				if i == 0 && j == 0 {
					continue
				}
				if relativePos.Y+i < 0 || relativePos.Y+i >= len(grid.CollisionGrid) || relativePos.X+j < 0 || relativePos.X+j >= len(grid.CollisionGrid[relativePos.Y]) {
					continue
				}
				if grid.CollisionGrid[relativePos.Y+i][relativePos.X+j] == game.CollisionTypeWalkable {
					grid.CollisionGrid[relativePos.Y+i][relativePos.X+j] = game.CollisionTypeLowPriority

				}
			}
		}
	}

	// Add monsters to the collision grid as obstacles
	for _, m := range pf.data.Monsters {
		if !grid.IsWalkable(m.Position) {
			continue
		}
		relativePos := grid.RelativePosition(m.Position)
		grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeMonster
	}

	path, distance, found := astar.CalculatePath(grid, from, to)

	if config.Koolo.Debug.RenderMap {
		pf.renderMap(grid, from, to, path)
	}

	return path, distance, found
}

func (pf *PathFinder) mergeGrids(to data.Position) (*game.Grid, error) {
	for _, a := range pf.data.AreaData.AdjacentLevels {
		destination := pf.data.Areas[a.Area]
		if destination.IsInside(to) {
			origin := pf.data.AreaData

			endX1 := origin.OffsetX + len(origin.Grid.CollisionGrid[0])
			endY1 := origin.OffsetY + len(origin.Grid.CollisionGrid)
			endX2 := destination.OffsetX + len(destination.Grid.CollisionGrid[0])
			endY2 := destination.OffsetY + len(destination.Grid.CollisionGrid)
			//pf.logger.Debug(fmt.Sprintf("Origin OffsetX: %d, OffsetY: %d", origin.OffsetX, origin.OffsetY))
			//pf.logger.Debug(fmt.Sprintf("Destination OffsetX: %d, OffsetY: %d", destination.OffsetX, destination.OffsetY))
			//pf.logger.Debug(fmt.Sprintf("Origin Width: %d, Height: %d", len(origin.Grid.CollisionGrid[0]), len(origin.Grid.CollisionGrid)))
			//pf.logger.Debug(fmt.Sprintf("Destination Width: %d, Height: %d", len(destination.Grid.CollisionGrid[0]), len(destination.Grid.CollisionGrid)))

			// Calculate the minimum and maximum X and Y coordinates
			minX := min(origin.OffsetX, destination.OffsetX)
			minY := min(origin.OffsetY, destination.OffsetY)
			maxX := max(endX1, endX2)
			maxY := max(endY1, endY2)

			// Calculate the dimensions of the merged grid
			width := maxX - minX
			height := maxY - minY
			//pf.logger.Debug(fmt.Sprintf("Merged grid width: %d, merged grid height: %d", width, height))

			// Create the result grid with the calculated dimensions
			resultGrid := make([][]game.CollisionType, height)
			for i := range resultGrid {
				resultGrid[i] = make([]game.CollisionType, width)
			}

			// Copy the origin and destination grids into the result grid with the correct offsets
			copyGrid(resultGrid, origin.CollisionGrid, origin.OffsetX-minX, origin.OffsetY-minY)
			copyGrid(resultGrid, destination.CollisionGrid, destination.OffsetX-minX, destination.OffsetY-minY)

			// Adjust the Y-coordinate of the destination grid
			for y := 0; y < len(destination.CollisionGrid); y++ {
				for x := 0; x < len(destination.CollisionGrid[0]); x++ {
					resultGrid[destination.OffsetY-minY+y][destination.OffsetX-minX+x] = destination.CollisionGrid[y][x]
				}
			}

			grid := game.NewGrid(resultGrid, minX, minY)

			return grid, nil
		}
	}

	return nil, fmt.Errorf("destination grid not found")
}

func copyGrid(dest [][]game.CollisionType, src [][]game.CollisionType, offsetX, offsetY int) {
	for y := 0; y < len(src); y++ {
		for x := 0; x < len(src[0]); x++ {
			dest[offsetY+y][offsetX+x] = src[y][x]
		}
	}
}

func (pf *PathFinder) GetClosestWalkablePath(dest data.Position) (Path, int, bool) {
	return pf.GetClosestWalkablePathFrom(pf.data.PlayerUnit.Position, dest)
}

func (pf *PathFinder) GetClosestWalkablePathFrom(from, dest data.Position) (Path, int, bool) {
	a := pf.data.AreaData
	if a.IsWalkable(dest) || !a.IsInside(dest) {
		slog.Any("PathFinder.GetClosestWalkablePathFrom", "Destination is already walkable or outside the area")
		path, distance, found := pf.GetPath(dest)
		if found {
			return path, distance, found
		}
	}

	maxRange := 20
	step := 4
	dst := 1

	for dst < maxRange {
		for i := -dst; i < dst; i += 1 {
			for j := -dst; j < dst; j += 1 {
				if math.Abs(float64(i)) >= math.Abs(float64(dst)) || math.Abs(float64(j)) >= math.Abs(float64(dst)) {
					cgY := dest.Y - pf.data.AreaOrigin.Y + j
					cgX := dest.X - pf.data.AreaOrigin.X + i
					if cgX > 0 && cgY > 0 && a.Height > cgY && a.Width > cgX && a.CollisionGrid[cgY][cgX] == game.CollisionTypeWalkable {
						return pf.GetPathFrom(from, data.Position{
							X: dest.X + i,
							Y: dest.Y + j,
						})
					}
				}
			}
		}
		dst += step
	}

	return nil, 0, false
}

func (pf *PathFinder) findNearbyWalkablePosition(target data.Position) (data.Position, bool) {
	// Search in expanding squares around the target position
	for radius := 1; radius <= 3; radius++ {
		for x := -radius; x <= radius; x++ {
			for y := -radius; y <= radius; y++ {
				if x == 0 && y == 0 {
					continue
				}
				pos := data.Position{X: target.X + x, Y: target.Y + y}
				if pf.data.AreaData.IsWalkable(pos) {
					return pos, true
				}
			}
		}
	}
	return data.Position{}, false
}
