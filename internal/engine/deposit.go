package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

func Deposit(ctx context.Context, r *actions.Runner, character string) error {
	maps, err := r.GetMaps(ctx, client.Bank)
	if err != nil {
		return fmt.Errorf("failed to get maps %w", err)
	}

	bankCoords := models.Coords{}
	for _, m := range maps {
		if m.Code == "bank" {
			bankCoords.X = m.X
			bankCoords.Y = m.Y
			break
		}
	}
	slog.Info("bank found", "coords", bankCoords)

	// get all character info
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character %w", err)
	}

	// goto bank if not there
	if c.X != bankCoords.X && c.Y != bankCoords.Y {
		m, err := r.Move(ctx, character, bankCoords.X, bankCoords.Y)
		if err != nil {
			return fmt.Errorf("failed to move %w", err)
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Info("moved to bank", "char", character, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}

	// we know the inventory once
	// deposit all
	for _, i := range *c.Inventory {
		if i.Quantity > 0 && i.Code != "" {
			bankresp, err := r.Deposit(ctx, character, i.Code, i.Quantity)
			if err != nil {
				return fmt.Errorf("failed to deposit %w", err)
			}
			cooldown := time.Until(bankresp.CooldownSchema.Expiration)
			slog.Info("deposited items into bank", "item", bankresp.Item, "cooldown", cooldown)
			time.Sleep(cooldown)
		}
		slog.Info("skipping", "slot", i.Slot)
	}

	return nil
}
