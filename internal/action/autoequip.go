package action

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/utils"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
)

var bodyloc = map[string][]string{
	// Class Specific
	//"abow": []string{"larm"}, // Amazon bows
	//"ajav": []string{"larm"}, // Amazon javs
	//"aspe": []string{"larm"}, // Amazon spears
	//"ashd": []string{"rarm"}, // Auric shields
	//"h2h":  []string{"larm"}, // Assasin claws
	//"h2h2": []string{"larm"}, // Assassin claws
	//"head": []string{"rarm"}, // Zombie heads
	"orb": []string{"larm"}, // Sorc orbs
	//"pelt": []string{"tors"}, // Druid pelts
	//"phlm": []string{"head"}, // Barb helms

	// Special cases with multiple slots
	"ring": []string{"lrin", "rrin"},

	// All 2h weapons and shields assigned to rarm
	"spea": []string{"rarm"},
	"staf": []string{"rarm"},
	"pole": []string{"rarm"},
	"shie": []string{"rarm"},

	// All other weapons assigned to larm
	"swor": []string{"larm"},
	"axe":  []string{"larm"},
	"club": []string{"larm"},
	"hamm": []string{"larm"},
	"jave": []string{"larm"},
	"knif": []string{"larm"},
	"mace": []string{"larm"},
	"scep": []string{"larm"},
	"wand": []string{"larm"},

	//Everything else
	"circ": []string{"head"},
	"helm": []string{"head"},
	"amul": []string{"neck"},
	"tors": []string{"tors"},
	"glov": []string{"glov"},
	"belt": []string{"belt"},
	"boot": []string{"feet"},

	//Excluding items that have limited uses (bows, throwing items etc)
	//"xbow": []string{"larm"},
	//"tpot": []string{"larm"},
	//"tkni": []string{"larm"},
	//"taxe": []string{"larm"},
	//"bow":  []string{"larm"},
	//"abow": []string{"larm"},
}

type BestItem struct {
	Item     data.Item
	Score    float64
	Location item.Location
	Used     bool
}

func isEquippable(i data.Item) bool {
	ctx := context.Get()
	//Probably a neater way of doing this
	if _, maxDurabilityFound := i.FindStat(stat.MaxDurability, 0); maxDurabilityFound || i.Desc().Type == "amul" || i.Desc().Type == "ring" {
		str, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.Strength, 0)
		dex, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.Dexterity, 0)
		lvl, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.Level, 0)

		if i.Identified && str.Value >= i.Desc().RequiredStrength && dex.Value >= i.Desc().RequiredDexterity && lvl.Value >= i.Desc().RequiredLevel {
			return true
		}
	}
	return false
}

func EvaluateAllItems() error {
	ctx := context.Get()
	bestItems := make(map[string]BestItem)
	usedItems := make(map[data.UnitID]bool)

	for _, it := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		score := TierScore(it)
		ctx.Logger.Debug(fmt.Sprintf("Equipped %s (Score: %.2f) found in %s", it.Name, score, it.Location.LocationType))
	}

	//if !ctx.Data.OpenMenus.Stash {
	//	bank, _ := ctx.Data.Objects.FindOne(object.Bank)
	//	err := InteractObject(bank, func() bool {
	//		return ctx.Data.OpenMenus.Stash
	//	})
	//	if err != nil {
	//		return err
	//	}
	//}
	// Get all items from inventory and stash
	allItems := ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash, item.LocationInventory, item.LocationEquipped)

	// Find best items for each slot
	for _, i := range allItems {
		if !isEquippable(i) {
			continue
		}

		// Get location for item
		location := bodyloc[i.Desc().Type]

		// Calculate score
		score := TierScore(i)

		for _, loc := range location {
			// Skip if item already used in another slot (except rings)
			if _, isUsed := usedItems[i.UnitID]; isUsed {
				continue
			}
			// Update best item if score is higher
			if current, exists := bestItems[loc]; !exists || score > current.Score {

				if exists {
					delete(usedItems, current.Item.UnitID)
				}

				bestItems[loc] = BestItem{
					Item:     i,
					Score:    score,
					Location: i.Location,
				}
				usedItems[i.UnitID] = true
			}
		}

	}
	for loc, i := range bestItems {
		if i.Location.LocationType == item.LocationEquipped {
			continue
		}

		ctx.Logger.Debug(fmt.Sprintf("Slot %s: %s (Score: %.2f) found in %s",
			loc, i.Item.Name, i.Score, i.Location.LocationType))

		equip(i.Item, loc)
	}
	//ctx.Logger.Debug(fmt.Sprintf("Best items: %v", bestItems))
	return nil
}

func equip(i data.Item, bodyloc string) error {
	// passing in bodyloc as a parameter cos rings have 2 locations
	ctx := context.Get()

	coords := getEquipCoords(bodyloc)
	ctx.Logger.Debug(fmt.Sprintf("Equipping %s to %s", i.Name, coords))
	//dequip(bodyloc)
	if i.Location.LocationType == item.LocationStash || i.Location.LocationType == item.LocationSharedStash {
		OpenStash()
	}
	newc := ui.GetScreenCoordsForItem(i)
	context.Get().HID.Click(game.LeftButton, newc.X, newc.Y)
	time.Sleep(500 * time.Millisecond)
	context.Get().HID.Click(game.LeftButton, coords.X, coords.Y)
	step.CloseAllMenus()
	DropMouseItem()
	return nil
}

func dequip(bodyloc string) error {
	ctx := context.Get()

	eqItems := ctx.Data.Inventory.ByLocation(item.LocationEquipped)
	step.CloseAllMenus()
	for _, i := range eqItems {
		if i.Desc().Type == bodyloc {
			// Check if the inventory is open, if not open it
			if !ctx.Data.OpenMenus.Inventory {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
			}

			// Wait a second
			utils.Sleep(1000)
			coords := getEquipCoords(bodyloc)
			ctx.Logger.Debug(fmt.Sprintf("Dequipping %s from %s", i.Name, coords))
			context.Get().HID.Click(game.LeftButton, coords.X, coords.Y)
			//DropInventoryItem(i)
			time.Sleep(500 * time.Millisecond)
			step.CloseAllMenus()
			time.Sleep(500 * time.Millisecond)
			DropMouseItem()

			return step.CloseAllMenus()
		}
	}

	return nil
}

func equipnew(itm data.Item) error {
	screenPos := ui.GetScreenCoordsForItem(itm)
	ctx := context.Get()
	ctx.Logger.Debug(fmt.Sprintf("Dequipping from %s", screenPos))

	context.Get().HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.ShiftKey)
	time.Sleep(500 * time.Millisecond)
	return nil
}

func getEquipCoords(bodyLoc string) data.Position {
	ctx := context.Get()
	if ctx.Data.LegacyGraphics {
		coord := ui.ClassicCoords[bodyLoc]
		return coord
	}
	coord := ui.ResurrectedCoords[bodyLoc]
	return coord
}
