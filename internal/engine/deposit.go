package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// DepositAll is an engine operation which commands a character
// to visit a bank and deposit all of their inventory
func DepositAll(ctx context.Context, r *actions.Runner, character string) error {
	maps, err := r.GetMaps(ctx, client.Bank)
	if err != nil {
		slog.Error("failed to get maps", "error", err)
		return err
	}

	bankCoords := models.Coords{}
	for _, m := range maps {
		if m.Code == "bank" {
			bankCoords.X = m.Coords.X
			bankCoords.Y = m.Coords.Y
			break
		}
	}
	slog.Debug("bank found", "coords", bankCoords)

	// get all character info
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Error("failed to get character", "error", err)
		return err
	}

	// goto bank if not there
	if c.X != bankCoords.X && c.Y != bankCoords.Y {
		m, err := r.Move(ctx, character, bankCoords.X, bankCoords.Y)
		if err != nil {
			slog.Error("failed to move", "error", err)
			return err
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Debug("moved to bank", "char", character, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}

	// we know the inventory once
	// deposit all
	for _, i := range *c.Inventory {
		if i.Quantity > 0 && i.Code != "" {
			bankresp, err := r.Deposit(ctx, character, i.Code, i.Quantity)
			if err != nil {
				slog.Error("failed to get deposit", "error", err)
				return err
			}
			cooldown := time.Until(bankresp.CooldownSchema.Expiration)
			slog.Debug("deposited item(s) into bank", "item", bankresp.Item, "qty", i.Quantity, "cooldown", cooldown)
			time.Sleep(cooldown)
		}
	}
	slog.Debug("deposit finished")

	return nil
}
