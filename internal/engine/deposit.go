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

	// navigate to bank if not already there
	if c.X != bankCoords.X || c.Y != bankCoords.Y {
		m, mErr := r.Move(ctx, character, bankCoords.X, bankCoords.Y)
		if mErr != nil {
			slog.Error("failed to move", "error", mErr)
			return mErr
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Debug("moved to bank", "char", character, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}

	// loop over the inventory and deposit all
	for _, i := range *c.Inventory {
		if i.Quantity > 0 && i.Code != "" {
			b, bErr := r.Deposit(ctx, character, i.Code, i.Quantity)
			if bErr != nil {
				slog.Error("failed to deposit", "error", bErr)
				return bErr
			}
			cooldown := time.Until(b.CooldownSchema.Expiration)
			slog.Info("deposited item into bank", "item", b.Item, "qty", i.Quantity, "cooldown", cooldown)
			time.Sleep(cooldown)
		}
	}
	slog.Debug("deposit finished")

	return nil
}
