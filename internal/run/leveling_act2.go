package run

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func (a Leveling) act2() error {
	running := false

	if running || a.ctx.Data.PlayerUnit.Area != area.LutGholein {
		return nil
	}

	running = true
	if a.ctx.Data.CharacterCfg.Game.Difficulty == difficulty.Normal && a.ctx.CharacterCfg.Character.UseMerc == false {
		a.ctx.CharacterCfg.Character.UseMerc = true
		action.ReviveMerc() // in case we already have one, it won't be replaced
		action.HireMerc()
	}

	for _, it := range a.ctx.Data.Inventory.ByLocation(item.LocationEquipped) {

		a.ctx.Logger.Debug(fmt.Sprintf("Equipped %v type %v, Desc %v, From Q %v", it, it.Type(), it.Desc(), it.IsFromQuest()))
	}

	if a.ctx.Data.Quests[quest.Act2RadamentsLair].HasStatus(quest.StatusQuestNotStarted) && a.ctx.Data.Quests[quest.Act2TheHoradricStaff].HasStatus(quest.StatusQuestNotStarted) && a.ctx.Data.Quests[quest.Act2TaintedSun].HasStatus(quest.StatusQuestNotStarted) && a.ctx.Data.Quests[quest.Act2TheSummoner].HasStatus(quest.StatusQuestNotStarted) && a.ctx.Data.Quests[quest.Act2ArcaneSanctuary].HasStatus(quest.StatusQuestNotStarted) && a.ctx.Data.Quests[quest.Act2TheSevenTombs].HasStatus(quest.StatusQuestNotStarted) {
		a.ctx.Logger.Debug("Talking to Jerhyn")
		action.InteractNPC(npc.Jerhyn)
	}

	// Find Horadric Cube
	_, found := a.ctx.Data.Inventory.Find("HoradricCube", item.LocationInventory, item.LocationStash)
	if found {
		a.ctx.Logger.Info("Horadric Cube found, skipping quest")
	} else {
		a.ctx.Logger.Info("Horadric Cube not found, starting quest")
		return a.findHoradricCube()
	}

	if a.ctx.Data.Quests[quest.Act2TheSummoner].Completed() {
		// Try to get level 21 before moving to Duriel and Act3

		if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 18 {
			return TalRashaTombs{}.Run()
		}

		return a.duriel()
	}

	_, horadricStaffFound := a.ctx.Data.Inventory.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)

	// Find Staff of Kings
	_, found = a.ctx.Data.Inventory.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if found || horadricStaffFound {
		a.ctx.Logger.Info("StaffOfKings found, skipping quest")
	} else {
		a.ctx.Logger.Info("StaffOfKings not found, starting quest")
		return a.findStaff()
	}

	// Find Amulet
	_, found = a.ctx.Data.Inventory.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if found || horadricStaffFound {
		a.ctx.Logger.Info("Amulet of the Viper found, skipping quest")
	} else {
		a.ctx.Logger.Info("Amulet of the Viper not found, starting quest")
		return a.findAmulet()
	}

	// Summoner
	a.ctx.Logger.Info("Starting summoner quest")
	//action.InteractNPC(npc.Drognan)
	//return error
	return a.summoner()

}

func (a Leveling) findHoradricCube() error {
	err := action.WayPoint(area.HallsOfTheDeadLevel2)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.HallsOfTheDeadLevel3)
	if err != nil {
		return err
	}
	action.ClearCurrentLevel(true, data.MonsterAnyFilter())

	return nil
}

func (a Leveling) findStaff() error {
	err := action.WayPoint(area.FarOasis)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.MaggotLairLevel1)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.MaggotLairLevel2)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.MaggotLairLevel3)
	if err != nil {
		return err
	}
	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
		if found {
			a.ctx.Logger.Info("Staff Of Kings chest found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.StaffOfKingsChest)
	if !found {
		return err
	}

	return action.InteractObject(obj, func() bool {
		for _, obj := range a.ctx.Data.Objects {
			if obj.Name == object.StaffOfKingsChest && !obj.Selectable {
				return true
			}
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) findAmulet() error {
	err := action.WayPoint(area.LostCity)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.ValleyOfSnakes)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.ClawViperTempleLevel1)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.ClawViperTempleLevel2)
	if err != nil {
		return err
	}
	err = action.MoveTo(func() (data.Position, bool) {
		chest, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
		if found {
			a.ctx.Logger.Info("Tainted Sun Altar found, moving to that room")
			return chest.Position, true
		}
		return data.Position{}, false
	})
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(15, data.MonsterAnyFilter())

	obj, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
	if !found {
		return err
	}

	err = action.InteractObject(obj, func() bool {
		updatedObj, found := a.ctx.Data.Objects.FindOne(object.TaintedSunAltar)
		if found {
			if !updatedObj.Selectable {
				a.ctx.Logger.Debug("Interacted with Tainted Sun Altar")
			}
			return !updatedObj.Selectable
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) summoner() error {
	// Start Summoner run to find and kill Summoner
	err := Summoner{}.Run()
	if err != nil {
		return err
	}

	tome, found := a.ctx.Data.Objects.FindOne(object.YetAnotherTome)
	if !found {
		return err
	}

	// Try to use the portal and discover the waypoint
	err = action.InteractObject(tome, func() bool {
		_, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
		return found
	})
	if err != nil {
		return err
	}

	portal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return err
	}

	err = action.InteractObject(portal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.CanyonOfTheMagi && a.ctx.Data.AreaData.IsInside(a.ctx.Data.PlayerUnit.Position)
	})
	if err != nil {
		return err
	}

	err = action.DiscoverWaypoint()
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Atma)
	if err != nil {
		return err
	}

	a.ctx.HID.PressKey(win.VK_ESCAPE)

	return nil
}

func (a Leveling) prepareStaff() error {
	horadricStaff, found := a.ctx.Data.Inventory.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if found {
		a.ctx.Logger.Info("Horadric Staff found!")
		if horadricStaff.Location.LocationType == item.LocationStash {
			a.ctx.Logger.Info("It's in the stash, let's pick it up")

			bank, found := a.ctx.Data.Objects.FindOne(object.Bank)
			if !found {
				a.ctx.Logger.Info("bank object not found")
			}

			err := action.InteractObject(bank, func() bool {
				return a.ctx.Data.OpenMenus.Stash
			})
			if err != nil {
				return err
			}

			screenPos := ui.GetScreenCoordsForItem(horadricStaff)
			a.ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
			utils.Sleep(300)
			a.ctx.HID.PressKey(win.VK_ESCAPE)

			return nil
		}
	}

	staff, found := a.ctx.Data.Inventory.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if !found {
		a.ctx.Logger.Info("Staff of Kings not found, skipping")
		return nil
	}

	amulet, found := a.ctx.Data.Inventory.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
	if !found {
		a.ctx.Logger.Info("Amulet of the Viper not found, skipping")
		return nil
	}

	err := action.CubeAddItems(staff, amulet)
	if err != nil {
		return err
	}

	err = action.CubeTransmute()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) duriel() error {
	a.ctx.Logger.Info("Starting Duriel....")
	if err := NewDuriel().Run(); err != nil {
		return fmt.Errorf("failed to complete Duriel: %w", err)
	}
	durielstatus := a.ctx.Data.Quests[quest.Act2TheSevenTombs]
	a.ctx.Logger.Debug(fmt.Sprintf("Duriel status: %v", durielstatus))
	utils.Sleep(500)
	action.MoveToCoords(data.Position{
		X: 22577,
		Y: 15613,
	})

	action.InteractNPC(npc.Tyrael)
	//a.ctx.HID.PressKey(win.VK_ESCAPE)
	durielstatus = a.ctx.Data.Quests[quest.Act2TheSevenTombs]
	a.ctx.Logger.Debug(fmt.Sprintf("Duriel status after Tyrael: %v", durielstatus))
	action.ReturnTown()
	action.MoveToCoords(data.Position{
		X: 5092,
		Y: 5144,
	})

	action.InteractNPC(npc.Jerhyn)
	//a.ctx.HID.PressKey(win.VK_ESCAPE)
	durielstatus = a.ctx.Data.Quests[quest.Act2TheSevenTombs]
	a.ctx.Logger.Debug(fmt.Sprintf("Duriel status after Jerhyn: %v", durielstatus))
	action.MoveToCoords(data.Position{
		X: 5195,
		Y: 5060,
	})
	action.InteractNPC(npc.Meshif)
	a.ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

	return nil
}
