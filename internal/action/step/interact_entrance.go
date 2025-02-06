package step

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
	"time"
)

const (
	maxEntranceDistance = 6
	maxMoveRetries      = 3
)

func InteractEntrance(targetArea area.ID) error {
	maxInteractionAttempts := 5
	interactionAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	// Find the closest entrance(if there are 2 entrances for same destination like harem/palace cellar)
	targetLevel := findClosestEntrance(ctx, targetArea)
	if targetLevel == nil {
		return fmt.Errorf("no entrance found for area %s [%d]", targetArea.Area().Name, targetArea)
	}

	for {
		ctx.PauseIfNotPriority()

		if ctx.Data.AreaData.Area == targetArea && time.Since(lastRun) > time.Millisecond*500 && ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
			return nil
		}

		if interactionAttempts > maxInteractionAttempts {
			return fmt.Errorf("area %s [%d] could not be interacted", targetArea.Area().Name, targetArea)
		}

		if waitingForInteraction && time.Since(lastRun) < time.Millisecond*500 {
			continue
		}

		lastRun = time.Now()
		for _, l := range ctx.Data.AdjacentLevels {
			if l.Area == targetArea {
				distance := ctx.PathFinder.DistanceFromMe(l.Position)
				if distance > maxEntranceDistance {
					// Try to move closer with retries
					for retry := 0; retry < maxMoveRetries; retry++ {
						if err := MoveTo(l.Position); err != nil {
							// If MoveTo fails, try direct movement
							screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(
								l.Position.X-2,
								l.Position.Y-2,
							)
							ctx.HID.Click(game.LeftButton, screenX, screenY)
							utils.Sleep(800)
							ctx.RefreshGameData()
						}

						// Check if we're close enough now
						newDistance := ctx.PathFinder.DistanceFromMe(l.Position)
						if newDistance <= maxEntranceDistance {
							break
						}

						if retry == maxMoveRetries-1 {
							return fmt.Errorf("entrance too far away (distance: %d)", distance)
						}
					}
				}

				if l.IsEntrance {
					lx, ly := ctx.PathFinder.GameCoordsToScreenCords(l.Position.X-1, l.Position.Y-1)
					ctx.Logger.Debug("Hovered over entrance, clicking")
					
					ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
					waitingForInteraction = true
					utils.Sleep(200)

					x, y := utils.Spiral(interactionAttempts)
					x = x / 3
					y = y / 3
					currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
					ctx.HID.MovePointer(lx+x, ly+y)
					interactionAttempts++
					utils.Sleep(100)
					continue
				}

				return fmt.Errorf("area %s [%d] is not an entrance", targetArea.Area().Name, targetArea)
			}
		}
	}
}

func findClosestEntrance(ctx *context.Status, targetArea area.ID) *data.Level {
	var closest *data.Level
	shortestDistance := 999999
	for _, l := range ctx.Data.AdjacentLevels {
		if l.Area == targetArea && l.IsEntrance {
			distance := ctx.PathFinder.DistanceFromMe(l.Position)
			if distance < shortestDistance {
				shortestDistance = distance
				lvl := l
				closest = &lvl
			}
		}
	}
	return closest
}
