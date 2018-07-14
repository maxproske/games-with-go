package ui2d

import (
	"github.com/maxproske/games-with-go/38_equipment/game"
	"github.com/veandco/go-sdl2/sdl"
)

func (ui *ui) CheckDroppedItem() *game.Item {
	invRect := ui.getInventoryRect()
	mousePos := ui.currentMouseState.pos
	// Rect of size 1
	if invRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
		return nil // Dropped in inventory rect
	}
	return ui.draggedItem // Dropped outside inventory rect
}

func (ui *ui) getHelmetSlotRect() *sdl.Rect {
	invRect := ui.getInventoryRect()
	slotSize := int32(itemSizeRatio * float32(ui.winWidth) * 1.05) // Multiply for extra padding
	return &sdl.Rect{(invRect.X*2+invRect.W)/2 - slotSize/2, invRect.Y, slotSize, slotSize}
}

func (ui *ui) getWeaponSlotRect() *sdl.Rect {
	invRect := ui.getInventoryRect()
	slotSize := int32(itemSizeRatio * float32(ui.winWidth) * 1.05) // Multiply for extra padding
	yoffset := int32(float32(invRect.H) * 0.18)
	xoffset := int32(float32(invRect.W) * 0.18)
	return &sdl.Rect{invRect.X + xoffset, invRect.Y + yoffset, slotSize, slotSize}
}

func (ui *ui) getInventoryRect() *sdl.Rect {
	invWidth := int32(float32(ui.winWidth) * 0.4)
	invHeight := int32(float32(ui.winHeight) * 0.75)
	offsetX := (int32(ui.winWidth) - invWidth) / 2
	offsetY := (int32(ui.winHeight) - invHeight) / 2
	return &sdl.Rect{offsetX, offsetY, invWidth, invHeight}
}

func (ui *ui) getInventoryItemRect(i int) *sdl.Rect {
	invRect := ui.getInventoryRect()
	itemSize := int32(itemSizeRatio * float32(ui.winWidth))
	return &sdl.Rect{invRect.X + int32(i)*itemSize, invRect.Y + invRect.H - itemSize, itemSize, itemSize}
}

// DrawInventory ....
func (ui *ui) DrawInventory(level *game.Level) {

	// Enlarge player image
	playerSrcRect := ui.textureIndex[level.Player.Rune][0]
	invRect := ui.getInventoryRect()
	ui.renderer.Copy(ui.groundInventoryBackground, nil, invRect)
	offset := int32(float64(invRect.H) * 0.05) // Padding between inventory area and player

	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{invRect.X + invRect.X/4, invRect.Y + offset, invRect.W / 2, invRect.H / 2})
	ui.renderer.Copy(ui.slotBackground, nil, ui.getHelmetSlotRect())
	ui.renderer.Copy(ui.slotBackground, nil, ui.getWeaponSlotRect())
	// Render equipped helmet
	if level.Player.Helmet != nil {
		ui.renderer.Copy(ui.textureAtlas, &ui.textureIndex[level.Player.Helmet.Rune][0], ui.getHelmetSlotRect())
	}
	// Render equipped weapon
	if level.Player.Weapon != nil {
		ui.renderer.Copy(ui.textureAtlas, &ui.textureIndex[level.Player.Weapon.Rune][0], ui.getWeaponSlotRect())
	}

	// Render items in player inventory
	for i, item := range level.Player.Items {
		itemSrcRect := ui.textureIndex[item.Rune][0]

		if item == ui.draggedItem {
			itemSize := int32(itemSizeRatio * float32(ui.winWidth))
			ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, &sdl.Rect{int32(ui.currentMouseState.pos.X), int32(ui.currentMouseState.pos.Y), itemSize, itemSize})
		} else {
			ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, ui.getInventoryItemRect(i))
		}
	}
}

func (ui *ui) CheckInventoryItems(level *game.Level) *game.Item {
	if ui.currentMouseState.leftButton {
		// Dragged
		mousePos := ui.currentMouseState.pos
		for i, item := range level.Player.Items {
			itemRect := ui.getInventoryItemRect(i)
			// Check if current mouse position is in inventory rect
			intersection := itemRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) // Pass mouse as a single pixel rect
			if intersection {
				// Clicked inventory item
				return item
			}
		}
	}

	return nil
}

func (ui *ui) CheckGroundItems(level *game.Level) *game.Item {
	if !ui.currentMouseState.leftButton && ui.prevMouseState.leftButton {
		// Clicked
		items := level.Items[level.Player.Pos]
		mousePos := ui.currentMouseState.pos
		for i, item := range items {
			itemRect := ui.getGroundItemRect(i)
			// Check if current mouse position is in ground item rect
			intersection := itemRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) // Pass mouse as a single pixel rect
			if intersection {
				return item
			}
		}
	}
	return nil
}

// Check our
func (ui *ui) CheckEquippedItem() *game.Item {
	// Assume we have a dragged item already
	mousePos := ui.currentMouseState.pos
	if ui.draggedItem.Typ == game.Weapon {
		r := ui.getWeaponSlotRect()
		if r.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
			return ui.draggedItem // If we are equipping anything, return it
		}
	} else if ui.draggedItem.Typ == game.Helmet {
		r := ui.getHelmetSlotRect()
		if r.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
			return ui.draggedItem // If we are equipping anything, return it
		}
	}
	return nil
}
