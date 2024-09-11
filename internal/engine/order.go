package engine

import (
	"context"
	"fmt"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// ShouldFulfilOrder determines if this order is still relevant / should be fulfilled
// it's based upon the quantity on hand in bank, not counting items in flight
func ShouldFulfilOrder(ctx context.Context, r *actions.Runner, order models.SimpleItem) bool {
	// determine what's in the bank
	items, err := r.GetBankItems(ctx)
	if err != nil {
		return false
	}

	var bankItem models.SimpleItem
	for _, item := range items {
		if item.Code == order.Code {
			bankItem = item
			break
		}
	}

	logging.Get(ctx).Debug("")

	if bankItem.Quantity < order.Quantity {
		logging.Get(ctx).Debug("order quantity is greater than quantity on hand", "resource", order.Code, "required", order.Quantity, "on_hand", bankItem.Quantity)
		return true
	} else {
		return false
	}

}

// FulfilOrder will instruct the character to seek out and gather the required resource
func FulfilOrder(ctx context.Context, r *actions.Runner, character string, order models.SimpleItem) error {
	// determine all resources that drop the order
	resources, err := r.GetResourcesByDrop(ctx, order.Code)
	if err != nil {
		return fmt.Errorf("get resources by drop: %w", err)
	}

	// get character location
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("get character info: %w", err)
	}

	if c.ShouldBank() {
		dErr := DepositAll(ctx, r, character)
		if dErr != nil {
			return fmt.Errorf("failed to deposit all: %w", dErr)
		}
	}

	// goto the nearest resource
	gErr := Gather(ctx, r, character, resources[0])
	if gErr != nil {
		return fmt.Errorf("failed to gather resources: %w", gErr)
	} else {
		dErr := DepositAll(ctx, r, character)
		if dErr != nil {
			return fmt.Errorf("failed to deposit all: %w", dErr)
		}
	}

	return nil
}