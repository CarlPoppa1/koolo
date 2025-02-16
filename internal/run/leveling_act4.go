package run

import (
	"errors"
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Leveling) act4() error {
	running := false
	if running || a.ctx.Data.PlayerUnit.Area != area.ThePandemoniumFortress {
		return nil
	}

	running = true

	if !a.ctx.Data.Quests[quest.Act4TheFallenAngel].Completed() {
		a.izual()
	}
	if !a.ctx.Data.Quests[quest.Act4TerrorsEnd].Completed() {
		diabloRun := NewDiablo()
		err := diabloRun.Run()
		if err != nil {
			return err
		}
	} else {
		err := action.InteractNPC(npc.Tyrael2)
		if err != nil {
			return err
		}
		harrogathPortal, found := a.ctx.Data.Objects.FindOne(object.LastLastPortal)
		if !found {
			return errors.New("Portal to Harrogath not found")
		}

		err = action.InteractObject(harrogathPortal, func() bool {
			return a.ctx.Data.AreaData.Area == area.Harrogath && a.ctx.Data.AreaData.IsInside(a.ctx.Data.PlayerUnit.Position)
		})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (a Leveling) izual() error {

	a.ctx.Logger.Debug("Starting Izual")
	err := action.MoveToArea(area.OuterSteppes)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.PlainsOfDespair)
	if err != nil {
		return err
	}
	action.Buff()

	areaData := a.ctx.Data.Areas[area.PlainsOfDespair]
	izualNPC, found := areaData.NPCs.FindOne(npc.Izual)
	if !found || len(izualNPC.Positions) == 0 {
		a.ctx.Logger.Error("Izual not found")
		return err
	}
	a.ctx.Logger.Debug(fmt.Sprintf("Izual found at %v", izualNPC))
	// Move to the Summoner's position using the static coordinates from map data
	if err = action.MoveToCoords(izualNPC.Positions[0]); err != nil {
		a.ctx.Logger.Debug("Failed to move to Izual")
		return err
	}

	err = action.MoveTo(func() (data.Position, bool) {
		izual, found := a.ctx.Data.NPCs.FindOne(npc.Izual)
		if !found {
			return data.Position{}, false
		}

		return izual.Positions[0], true
	})
	if err != nil {
		return err
	}

	err = a.ctx.Char.KillIzual()
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Tyrael2)
	if err != nil {
		return err
	}

	return nil
}
