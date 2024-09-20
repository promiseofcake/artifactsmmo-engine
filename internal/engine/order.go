package engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// ShouldFulfilOrder determines if this order is still relevant / should be fulfilled
// it's based upon the quantity on hand in bank, not counting items in flight
func ShouldFulfilOrder(ctx context.Context, r *actions.Runner, c models.Character, order models.Order) bool {
	// determine what's in the bank
	items, err := r.GetBankItems(ctx)
	if err != nil {
		return false
	}

	var bankItem models.SimpleItem
	for _, item := range items {
		if item.Code == order.Item.Code {
			bankItem = item
			break
		}
	}

	var inventoryItem models.SimpleItem
	for _, slot := range *c.Inventory {
		if slot.Code == order.Item.Code {
			inventoryItem = models.SimpleItem{Code: slot.Code, Quantity: slot.Quantity}
			break
		}
	}

	if (bankItem.Quantity + inventoryItem.Quantity) < order.Item.Quantity {
		logging.Get(ctx).Debug("order quantity is greater than quantity on hand", "resource", order.Item.Code, "required", order.Item.Quantity, "inventory", inventoryItem.Quantity, "bank", bankItem.Quantity)
		return true
	} else {
		return false
	}

}

// FulfilOrder will instruct the character to seek out and gather the required resource
func FulfilOrder(ctx context.Context, r *actions.Runner, character string, order models.Order) ([]models.Order, error) {
	l := logging.Get(ctx)

	// get character location
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("get character info: %w", err)
	}

	// determine if it's a resource or a craft
	item, err := r.GetItem(ctx, order.Item.Code)
	l.Debug("get item details", "item", item)
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}

	// resource only
	if item.Craft == nil {
		l.Debug("this is a gather resource")
		// determine all resources that drop the order
		resources, err := r.GetResourcesByDrop(ctx, order.Item.Code)
		if err != nil {
			return nil, fmt.Errorf("get resources by drop: %w", err)
		}

		if c.ShouldBank() {
			dErr := DepositAll(ctx, r, character)
			if dErr != nil {
				return nil, fmt.Errorf("failed to deposit all: %w", dErr)
			}
		}

		// goto the nearest resource
		gErr := Gather(ctx, r, character, resources[0])
		if gErr != nil {
			return nil, fmt.Errorf("failed to gather resources: %w", gErr)
		}
	} else {
		l.Debug("this is a craft resource")
		cs, err := item.Craft.AsCraftSchema()
		if err != nil {
			return nil, fmt.Errorf("get item craft schema: %w", err)
		}
		// items

		var reqs []models.Order
		for _, input := range *cs.Items {

			io := models.Order{
				Item: models.SimpleItem{
					Code:     input.Code,
					Quantity: input.Quantity * order.Item.Quantity,
				},
				Concurrency: order.Concurrency,
				Action:      "gather",
			}

			if ShouldFulfilOrder(ctx, r, c, io) {
				l.Debug("missing required item for crafting", "order", io)
				reqs = append(reqs, io)
			}
		}

		if len(reqs) > 0 {
			return reqs, errors.New("requirements not met")
		}

		l.Debug("all items present for crafting!")

		// go to bank
		// withdraw items
		// craft items

		return nil, errors.New("failed to implemnet crafting")
	}

	return nil, nil
}
