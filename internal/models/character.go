package models

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Character is our representation of the Player's character
// Ideally we can update this state with the value of a response
// from a request.
type Character struct {
	client.CharacterSchema
}

// CountInventory returns the number of items in the Character's Inventory
func (c Character) CountInventory() int {
	var count int
	for _, item := range *c.Inventory {
		count += item.Quantity
	}
	return count
}

// GetCooldownDuration returns the time.Duration remaining on the character for cooldown
func (c Character) GetCooldownDuration() (time.Duration, error) {
	t, err := c.CooldownExpiration.AsCharacterSchemaCooldownExpiration0()
	if err != nil {
		return 0, fmt.Errorf("failed to get cooldown expiration: %w", err)
	}

	return time.Until(t), nil
}

// GetPosition returns the given Coordinates for the Character in question
func (c Character) GetPosition() Coords {
	return Coords{
		X: c.X,
		Y: c.Y,
	}
}

// ShouldBank will determine if the character should empty their inventory to the bank
func (c Character) ShouldBank() bool {
	percentFull := float64(c.CountInventory()) / float64(c.InventoryMaxItems)
	result := []any{"percent_full", percentFull}
	if percentFull > 0.7 {
		slog.Debug("Character should bank", result...)
		return true
	} else {
		slog.Debug("Character should not bank", result...)
		return false
	}
}
