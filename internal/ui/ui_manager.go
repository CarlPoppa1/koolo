package ui

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/koolo/internal/context"
)

var (
	ClassicCoords = map[string]data.Position{
		"head": {X: EquipHeadClassicX, Y: EquipHeadClassicY},
		"neck": {X: EquipNeckClassicX, Y: EquipNeckClassicY},
		"larm": {X: EquipLArmClassicX, Y: EquipLArmClassicY},
		"rarm": {X: EquipRArmClassicX, Y: EquipRArmClassicY},
		"tors": {X: EquipTorsClassicX, Y: EquipTorsClassicY},
		"belt": {X: EquipBeltClassicX, Y: EquipBeltClassicY},
		"glov": {X: EquipGlovClassicX, Y: EquipGlovClassicY},
		"feet": {X: EquipFeetClassicX, Y: EquipFeetClassicY},
		"lrin": {X: EquipLRinClassicX, Y: EquipLRinClassicY},
		"rrin": {X: EquipRRinClassicX, Y: EquipRRinClassicY},
	}

	ResurrectedCoords = map[string]data.Position{
		"head": {X: EquipHeadX, Y: EquipHeadY},
		"neck": {X: EquipNeckX, Y: EquipNeckY},
		"larm": {X: EquipLArmX, Y: EquipLArmY},
		"rarm": {X: EquipRArmX, Y: EquipRArmY},
		"tors": {X: EquipTorsX, Y: EquipTorsY},
		"belt": {X: EquipBeltX, Y: EquipBeltY},
		"glov": {X: EquipGlovX, Y: EquipGlovY},
		"feet": {X: EquipFeetX, Y: EquipFeetY},
		"lrin": {X: EquipLRinX, Y: EquipLRinY},
		"rrin": {X: EquipRRinX, Y: EquipRRinY},
	}
)

func GetScreenCoordsForItem(itm data.Item) data.Position {
	ctx := context.Get()
	if ctx.GameReader.LegacyGraphics() {
		return getScreenCoordsForItemClassic(itm)
	}

	return getScreenCoordsForItem(itm)
}

func getScreenCoordsForItem(itm data.Item) data.Position {
	switch itm.Location.LocationType {
	case item.LocationVendor, item.LocationStash, item.LocationSharedStash:
		x := topCornerVendorWindowX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
		y := topCornerVendorWindowY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

		return data.Position{X: x, Y: y}
	case item.LocationCube:
		x := topCornerCubeWindowX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
		y := topCornerCubeWindowY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

		return data.Position{X: x, Y: y}
	}

	x := inventoryTopLeftX + itm.Position.X*itemBoxSize + (itemBoxSize / 2)
	y := inventoryTopLeftY + itm.Position.Y*itemBoxSize + (itemBoxSize / 2)

	return data.Position{X: x, Y: y}
}

func getScreenCoordsForItemClassic(itm data.Item) data.Position {
	switch itm.Location.LocationType {
	case item.LocationVendor, item.LocationStash, item.LocationSharedStash:
		x := topCornerVendorWindowXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
		y := topCornerVendorWindowYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

		return data.Position{X: x, Y: y}
	case item.LocationCube:
		x := topCornerCubeWindowXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
		y := topCornerCubeWindowYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

		return data.Position{X: x, Y: y}
	}

	x := inventoryTopLeftXClassic + itm.Position.X*itemBoxSizeClassic + (itemBoxSizeClassic / 2)
	y := inventoryTopLeftYClassic + itm.Position.Y*itemBoxSizeClassic + (itemBoxSizeClassic / 2)

	return data.Position{X: x, Y: y}
}
