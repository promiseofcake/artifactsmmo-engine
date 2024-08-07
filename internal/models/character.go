package models

import (
	"log/slog"

	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Character is our representation of the Player's character
// Ideally we can update this state with the value of a response
// from a request.
type Character struct {
	client.CharacterSchema
}

// CountInventory returns the number of items in the Character's Inventory
func (p *Character) CountInventory() int {
	var count int
	for _, item := range *p.Inventory {
		count += item.Quantity
	}
	return count
}

// GetPosition returns the given Coords for the Character in question
func (c *Character) GetPosition() Coords {
	return Coords{
		X: c.X,
		Y: c.Y,
	}
}

// ShouldBank will determine if the character should empty their inventory to the bank
func (p *Character) ShouldBank() bool {
	percentFull := float64(p.CountInventory()) / float64(p.InventoryMaxItems)
	result := []any{"percent_full", percentFull}
	if percentFull > 0.7 {
		slog.Debug("Character should bank", result...)
		return true
	} else {
		slog.Debug("Character should not bank", result...)
		return false
	}
}
