package engine

import (
	"context"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// DepositAll is an engine operation which commands a character
// to visit a bank and deposit all of their inventory
func DepositAll(ctx context.Context, r actions.Runner, character string) error {
	l := logging.Get(ctx)
	err := Travel(ctx, r, character, models.Location{
		Type: string(client.Bank),
		Code: string(client.Bank),
	})
	if err != nil {
		l.Error("failed to travel to bank", "error", err)
		return err
	}

	// get all character info
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		l.Error("failed to get character", "error", err)
		return err
	}

	// loop over the inventory and deposit all
	for _, i := range *c.Inventory {
		if i.Quantity > 0 && i.Code != "" {
			b, bErr := r.Deposit(ctx, character, i.Code, i.Quantity)
			if bErr != nil {
				l.Error("failed to deposit", "error", bErr)
				return bErr
			}
			cooldown := time.Until(b.CooldownSchema.Expiration)
			l.Info("deposited item into bank", "item", b.Item, "qty", i.Quantity, "cooldown", cooldown)
			time.Sleep(cooldown)
		}
	}
	l.Debug("deposit finished")

	return nil
}
