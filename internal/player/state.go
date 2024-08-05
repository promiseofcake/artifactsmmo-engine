package player

import (
	"log/slog"

	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

type Character struct {
	*client.CharacterSchema
}

func (p *Character) CountInventory() int {
	var count int
	for _, item := range *p.Inventory {
		count += item.Quantity
	}
	return count
}

func (p *Character) ShouldBank() bool {
	percentFull := float64(p.CountInventory()) / float64(p.InventoryMaxItems)
	result := []any{"percent_full", percentFull}
	if percentFull > 0.7 {
		slog.Info("Character should bank", result...)
		return true
	} else {
		slog.Info("Character should not bank", result...)
		return false
	}
}
