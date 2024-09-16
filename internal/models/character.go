package models

import (
	"cmp"
	"log/slog"
	"slices"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/math"
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

type CharacterSkill struct {
	Code         client.ResourceSchemaSkill
	CurrentLevel int
	MinLevel     int
}

func (c Character) ChooseWeakestSkill() CharacterSkill {
	skills := []CharacterSkill{
		{
			Code:         client.ResourceSchemaSkillWoodcutting,
			CurrentLevel: c.WoodcuttingLevel,
			MinLevel:     map[bool]int{true: 20, false: math.Max(0, c.WoodcuttingLevel-10)}[c.WoodcuttingLevel == 35],
		},
		{
			Code:         client.ResourceSchemaSkillMining,
			CurrentLevel: c.MiningLevel,
			MinLevel:     map[bool]int{true: 20, false: math.Max(0, c.MiningLevel-10)}[c.MiningLevel == 35],
		},
		{
			Code:         client.ResourceSchemaSkillFishing,
			CurrentLevel: c.FishingLevel,
			MinLevel:     map[bool]int{true: 20, false: math.Max(0, c.FishingLevel-10)}[c.FishingLevel == 35],
		},
	}

	slices.SortFunc(skills, func(a, b CharacterSkill) int {
		return cmp.Compare(a.CurrentLevel, b.CurrentLevel)
	})
	return skills[0]
}

// GetCooldownDuration returns the time.Duration remaining on the character for cooldown
func (c Character) GetCooldownDuration() (time.Duration, error) {
	t := c.CooldownExpiration
	return time.Until(*t), nil
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
	l := slog.With("character", c.Name)
	percentFull := float64(c.CountInventory()) / float64(c.InventoryMaxItems)
	result := []any{"percent_full", percentFull}
	if percentFull > 0.9 {
		l.Debug("Character should bank", result...)
		return true
	} else {
		l.Debug("Character should not bank", result...)
		return false
	}
}
