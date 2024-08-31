package engine

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

var NoItemsToRefine = errors.New("no items to refine")

func Refine(ctx context.Context, r *actions.Runner, character string) error {
	var err error
	// get all bank items, determine what's available to refine
	banked, err := r.GetBankItems(ctx)
	if err != nil {
		slog.Debug("failed to get bank items", "error", err)
		return err
	}

	// given all the banked items, get information on each of them and call
	// them resources
	var resources models.Items
	for _, b := range banked {
		item, iErr := r.GetItem(ctx, b.Code)
		if iErr != nil {
			slog.Debug("failed to get item info", "error", iErr)
			return iErr
		}

		// TODO, make this move beyond cooking / fishing
		item.Quantity = b.Quantity
		if item.Type == "resource" && item.Subtype == string(client.Fishing) {
			resources = append(resources, item)
		}
	}

	// get characater state
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Debug("failed to get character", "error", err)
		return err
	}

	// given the resources we have (hydrated bank items), try to determine which of these items
	// our character is available to refine based upon their level
	var refinable models.Items
	for _, res := range resources {
		var refineLevel int
		var skillType string

		// given the resource's subtype, lookup this characters
		// skill level, and fetch all items one could make
		// with the current given skill level
		switch res.Subtype {
		case string(client.Woodcutting):
			refineLevel = c.WoodcuttingLevel
			skillType = string(client.CraftSchemaSkillWoodcutting)
		case string(client.Mining):
			refineLevel = c.MiningLevel
			skillType = string(client.CraftSchemaSkillMining)
		case string(client.Fishing):
			refineLevel = c.CookingLevel
			skillType = string(client.CraftSchemaSkillCooking)
		}

		items, iErr := r.GetItems(ctx, 0, refineLevel, skillType, res.Code)
		if iErr != nil {
			return iErr
		}
		refinable = append(refinable, items...)
	}

	if len(refinable) == 0 {
		return NoItemsToRefine
	}

	// since we know there are items we can refine
	// we need to go to the bank to refine them
	err = Travel(ctx, r, character, models.Location{
		Code: "bank",
		Type: "bank",
	})
	if err != nil {
		return err
	}

	// empty the inventory to maximize refining
	err = DepositAll(ctx, r, character)
	if err != nil {
		return err
	}

	// given the first item in the list (we should sort it)
	// determine how much we want to withdraw.
	slices.SortFunc(refinable, func(a, b models.Item) int {
		return cmp.Compare(b.Level, a.Level)
	})

	// look over all the resources we have in the bank
	// look over all the refinable items that match that
	// resource, and then go for the first one.
	var resourceToRefine = refinable[0]
	var available int
	for _, res := range resources {
		if res.Code == resourceToRefine.RawCode {
			available = res.Quantity
		}
	}
	qty := int(math.Min(float64(c.InventoryMaxItems), float64(available)))

	resp, err := r.Withdraw(ctx, character, resourceToRefine.Code, qty)
	if err != nil {
		return err
	}

	cooldown := time.Until(resp.CooldownSchema.Expiration)
	c.CharacterSchema = resp.CharacterResponse.CharacterSchema
	time.Sleep(cooldown)

	// need to travel to refinement location
	err = Travel(ctx, r, character, models.Location{
		// TODO, expand to be the type based upon the resource
		Code: "cooking",
		Type: "workshop",
	})
	if err != nil {
		return err
	}

	// need to refine the item
	skillresp, err := r.Craft(ctx, character, resourceToRefine.Code, qty)
	if err != nil {
		return err
	}
	slog.Debug("skill response", "response", skillresp)

	cooldown = time.Until(skillresp.Response.CooldownSchema.Expiration)
	c.CharacterSchema = skillresp.Response.CharacterResponse.CharacterSchema
	time.Sleep(cooldown)

	// need to return to bank and deposit
	err = DepositAll(ctx, r, character)
	if err != nil {
		return err
	}

	return nil
}
