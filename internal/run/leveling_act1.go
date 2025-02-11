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
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func (a Leveling) act1() error {
	running := false
	if running || a.ctx.Data.PlayerUnit.Area != area.RogueEncampment {
		return nil
	}

	running = true

	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value == 1 {
		a.ctx.CharacterCfg.Game.Difficulty = difficulty.Normal
		a.ctx.CharacterCfg.MaxGameLength = 750
		a.ctx.CharacterCfg.BackToTown.EquipmentBroken = false
		a.ctx.CharacterCfg.BackToTown.MercDied = false
		a.ctx.CharacterCfg.BackToTown.NoHpPotions = false
		a.ctx.CharacterCfg.BackToTown.NoMpPotions = false
		a.ctx.CharacterCfg.Inventory.InventoryLock = [][]int{
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		}
		a.ctx.CharacterCfg.Character.UseMerc = false
		a.ctx.CharacterCfg.Health.ChickenAt = 10
		a.ctx.CharacterCfg.Health.ManaPotionAt = 65
		a.ctx.CharacterCfg.Health.HealingPotionAt = 80
		a.ctx.CharacterCfg.Health.RejuvPotionAtLife = 50
		config.SaveSupervisorConfig(a.ctx.Name, a.ctx.CharacterCfg)
	}
	if a.ctx.Data.PlayerUnit.TotalPlayerGold() < 2000 {
		a.ctx.CharacterCfg.BackToTown.NoHpPotions = false
		a.ctx.CharacterCfg.BackToTown.NoMpPotions = false
	} else {
		a.ctx.CharacterCfg.BackToTown.NoHpPotions = true
		a.ctx.CharacterCfg.BackToTown.NoMpPotions = true
	}

	// clear den of evil
	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 3 {
		a.ctx.Logger.Debug("Current lvl %s under 3 - Leveling in Den of Evil")
		a.denOfEvil()
		return fmt.Errorf("den of Evil finished")
	}
	// do Cold Plains until level 7
	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 6 {
		a.coldPlains()
		return fmt.Errorf("Cold Plains finished")
	}

	// do Countess Runs until level 14
	if lvl, _ := a.ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 14 {
		a.countess()
		return fmt.Errorf("Countess finished")
	}

	if !a.isCainInTown() && !a.ctx.Data.Quests[quest.Act1TheSearchForCain].Completed() {
		a.deckardCain()
	}

	if a.ctx.Data.Quests[quest.Act1SistersToTheSlaughter].Completed() {
		action.ReturnTown()
		// Do Den of Evil if not complete before moving acts
		if !a.ctx.Data.Quests[quest.Act1DenOfEvil].Completed() {
			a.denOfEvil()
		}
		action.InteractNPC(npc.Warriv)
		a.ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

		return nil
	} else {
		return a.andariel()
	}
}

func (a Leveling) bloodMoor() error {
	err := action.MoveToArea(area.BloodMoor)
	if err != nil {
		return err
	}

	return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
}

func (a Leveling) coldPlains() error {
	err := action.MoveToArea(area.ColdPlains)
	if err != nil {
		return err
	}

	return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
}

func (a Leveling) denOfEvil() error {
	err := action.MoveToArea(area.BloodMoor)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.DenOfEvil)
	if err != nil {
		return err
	}

	action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	action.ReturnTown()
	action.InteractNPC(npc.Akara)
	a.ctx.HID.PressKey(win.VK_ESCAPE)

	return nil
}

func (a Leveling) stonyField() error {
	err := action.WayPoint(area.StonyField)
	if err != nil {
		return err
	}
	// Find the Cairn Stone Alpha
	cairnStone := data.Object{}
	for _, o := range a.ctx.Data.Objects {
		if o.Name == object.CairnStoneAlpha {
			cairnStone = o
		}
	}

	// Move to the cairnStone
	action.MoveToCoords(cairnStone.Position)

	return action.ClearAreaAroundPlayer(10, data.MonsterEliteFilter())

	//return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
}

func (a Leveling) isCainInTown() bool {
	_, found := a.ctx.Data.Monsters.FindOne(npc.DeckardCain5, data.MonsterTypeNone)

	return found
}

func (a Leveling) deckardCain() error {
	action.WayPoint(area.RogueEncampment)
	a.ctx.Logger.Debug(fmt.Sprintf("Current quest status: %d", quest.Status(quest.Act1TheSearchForCain)))
	if a.ctx.Data.Quests[quest.Act1TheSearchForCain].HasStatus(quest.StatusQuestNotStarted) {
		err := action.WayPoint(area.DarkWood)
		if err != nil {
			return err
		}

		err = action.MoveTo(func() (data.Position, bool) {
			for _, o := range a.ctx.Data.Objects {
				if o.Name == object.InifussTree {
					return o.Position, true
				}
			}
			return data.Position{}, false
		})
		if err != nil {
			return err
		}

		action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())

		obj, found := a.ctx.Data.Objects.FindOne(object.InifussTree)
		if !found {
			a.ctx.Logger.Debug("InifussTree not found")
		}

		err = action.InteractObject(obj, func() bool {
			updatedObj, found := a.ctx.Data.Objects.FindOne(object.InifussTree)
			if found {
				if !updatedObj.Selectable {
					a.ctx.Logger.Debug("Interacted with InifussTree")
				}
				return !updatedObj.Selectable
			}
			return false
		})
		if err != nil {
			return err
		}

		action.ItemPickup(0)
		action.ReturnTown()
		action.InteractNPC(npc.Akara)
		a.ctx.HID.PressKey(win.VK_ESCAPE)
	}
	//Reuse Tristram Run actions Run actions
	err := a.tristram()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) tristram() error {
	Tristram{}.Run()
	return nil
}

func (a Leveling) countess() error {
	a.ctx.Logger.Debug("Current lvl %s under 19 - Leveling in Countess")
	err := action.WayPoint(area.BlackMarsh)
	if err != nil {
		return err
	}

	areas := []area.ID{
		area.ForgottenTower,
		area.TowerCellarLevel1,
	}

	for _, a := range areas {
		err = action.MoveToArea(a)
		if err != nil {
			return err
		}
	}

	currentAreaID := a.ctx.Data.PlayerUnit.Area.Area().ID

	for currentAreaID != area.TowerCellarLevel5 {
		pos1 := func() data.Position {
			for _, al := range a.ctx.Data.AdjacentLevels {
				if al.Area > currentAreaID {
					return al.Position // Return the actual position
				}
			}
			return data.Position{} // Return zero value if not found
		}

		err := action.ClearThroughPath(pos1(), 10, data.MonsterAnyFilter())
		if err != nil {
			return err
		}

		moved := false
		for _, al := range a.ctx.Data.AdjacentLevels {
			if al.Area > currentAreaID {
				err = action.MoveToArea(al.Area.Area().ID)
				if err != nil {
					return err
				}
				currentAreaID = al.Area.Area().ID
				moved = true
				break
			}
		}

		if !moved {
			return fmt.Errorf("No adjacent level found to move to from area %d", currentAreaID)
		}
	}

	return action.ClearCurrentLevel(false, data.MonsterEliteFilter())
}

func (a Leveling) andariel() error {
	err := action.WayPoint(area.CatacombsLevel2)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.CatacombsLevel3)
	action.MoveToArea(area.CatacombsLevel4)
	if err != nil {
		return err
	}

	// Return to the city, ensure we have pots and everything, and get some antidote potions
	action.ReturnTown()

	potsToBuy := 4
	if a.ctx.Data.MercHPPercent() > 0 {
		potsToBuy = 8
	}

	action.VendorRefill(false, true)
	action.BuyAtVendor(npc.Akara, action.VendorItemRequest{
		Item:     "AntidotePotion",
		Quantity: potsToBuy,
		Tab:      4,
	})

	a.ctx.HID.PressKeyBinding(a.ctx.Data.KeyBindings.Inventory)

	x := 0
	for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if itm.Name != "AntidotePotion" {
			continue
		}
		pos := ui.GetScreenCoordsForItem(itm)
		utils.Sleep(500)

		if x > 3 {
			a.ctx.HID.Click(game.LeftButton, pos.X, pos.Y)
			utils.Sleep(300)
			if a.ctx.Data.LegacyGraphics {
				a.ctx.HID.Click(game.LeftButton, ui.MercAvatarPositionXClassic, ui.MercAvatarPositionYClassic)
			} else {
				a.ctx.HID.Click(game.LeftButton, ui.MercAvatarPositionX, ui.MercAvatarPositionY)
			}
		} else {
			a.ctx.HID.Click(game.RightButton, pos.X, pos.Y)
		}
		x++
	}

	a.ctx.HID.PressKey(win.VK_ESCAPE)

	action.UsePortalInTown()
	action.Buff()

	action.MoveTo(func() (data.Position, bool) {
		return andarielAttackPos1, true
	})
	a.ctx.Char.KillAndariel()
	action.ReturnTown()
	action.InteractNPC(npc.Warriv)
	a.ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)

	return nil
}
