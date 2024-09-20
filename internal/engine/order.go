package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// ShouldFulfilOrder determines if this order is still relevant / should be fulfilled
// it's based upon the quantity on hand in bank, not counting items in flight
func ShouldFulfilOrder(ctx context.Context, r *actions.Runner, c models.Character, order models.Order) bool {
	// refresh char data
	c, err := r.GetMyCharacterInfo(ctx, c.Name)
	if err != nil {
		return false
	}

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
		item.Skill = string(*cs.Skill)

		// items
		var newOrders []models.Order
		var materials []models.Order
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
				newOrders = append(newOrders, io)
			} else {
				materials = append(materials, io)
			}
		}

		if len(newOrders) > 0 {
			return newOrders, errors.New("requirements not met")
		}

		l.Debug("all items present for crafting!")

		// check inventory

		// go to bank

		tErr := Travel(ctx, r, character, models.Location{
			Code: "bank",
			Type: "bank",
		})
		if tErr != nil {
			return nil, tErr
		}

		// deposit all
		dErr := DepositAll(ctx, r, character)
		if dErr != nil {
			return nil, fmt.Errorf("failed to deposit all: %w", dErr)
		}

		// withdraw items
		for _, mat := range materials {
			l.Info("withdrawing item", "code", mat.Item.Code, "qty", mat.Item.Quantity)
			resp, wErr := r.Withdraw(ctx, character, mat.Item.Code, mat.Item.Quantity)
			if wErr != nil {
				return nil, wErr
			}
			cooldown := time.Until(resp.CooldownSchema.Expiration)
			c.CharacterSchema = resp.CharacterResponse.CharacterSchema
			time.Sleep(cooldown)
		}

		// craft items
		l.Info("traveling to workshop", "skill", item.Skill)
		err = Travel(ctx, r, character, models.Location{
			Code: item.Skill,
			Type: "workshop",
		})
		if err != nil {
			return nil, err
		}

		skillresp, err := r.Craft(ctx, character, order.Item.Code, order.Item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to craft %s, %d, code: %w", order.Item.Code, order.Item.Quantity, err)
		}
		l.Info("skill response", "response", skillresp.SkillInfo)

		cooldown := time.Until(skillresp.Response.CooldownSchema.Expiration)
		c.CharacterSchema = skillresp.Response.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)

		return nil, nil
	}

	return nil, nil
}
