package engine

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

var NoItemsToRefine = errors.New("no items to refine")

func Refine(ctx context.Context, r *actions.Runner, character string) error {
	var err error
	// need to determine what is available to refine
	banked, err := r.GetBankItems(ctx)
	if err != nil {
		slog.Debug("failed to get bank items", "error", err)
		return err
	}

	// lookup all the bank items and make a list of reosurces
	var resources models.Items
	for _, b := range banked {
		item, iErr := r.GetItem(ctx, b.Code)
		if iErr != nil {
			slog.Debug("failed to get item info", "error", iErr)
			return iErr
		}

		// only look for resources we can work with
		item.Quantity = b.Quantity
		if item.Type == "resource" && item.Subtype == string(client.Fishing) {
			resources = append(resources, item)
		}
	}

	// given a list of all the current resources and their quantity, determine what to refine
	// subtype is the thing itself
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Debug("failed to get character", "error", err)
		return err
	}

	// we know they are things to be cooked
	// so we want to get the cooked versions of these things and determine which we can cook

	// what to refine?
	var items models.Items
	for _, have := range resources {
		// given subtype, lookup level
		var cl int
		var st string
		switch have.Subtype {
		//case string(client.Woodcutting):
		//	cl = c.WoodcuttingLevel
		//case string(client.Mining):
		//	cl = c.MiningLevel
		case string(client.Fishing):
			cl = c.CookingLevel
			st = "cooking"
		}

		ii, iErr := r.GetItems(ctx, 0, cl, st, have.Code)
		if iErr != nil {
			return iErr
		}
		items = append(items, ii...)
	}

	if len(items) == 0 {
		return NoItemsToRefine
	}

	// need to travel to bank
	err = Travel(ctx, r, character, models.Location{
		Code: "bank",
		Type: "bank",
	})
	if err != nil {
		return err
	}

	err = DepositAll(ctx, r, character)
	if err != nil {
		return err
	}

	// find item and quantity to withdraw
	var getQ int
	for _, get := range resources {
		if get.Code == items[0].RawCode {
			getQ = get.Quantity
		}
	}

	// need to withdraw
	fetched := int(math.Min(float64(c.InventoryMaxItems), float64(getQ)))
	resp, err := r.Client.ActionWithdrawBankMyNameActionBankWithdrawPostWithResponse(ctx, c.Name, client.ActionWithdrawBankMyNameActionBankWithdrawPostJSONRequestBody{
		Code:     items[0].RawCode,
		Quantity: fetched,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New(resp.Status())
	}

	cooldown := time.Until(resp.JSON200.Data.Cooldown.Expiration)
	c.CharacterSchema = resp.JSON200.Data.Character
	time.Sleep(cooldown)

	// need to travel to refinement location
	err = Travel(ctx, r, character, models.Location{
		Code: "cooking",
		Type: "workshop",
	})
	if err != nil {
		return err
	}

	// need to refine
	skillresp, err := r.Craft(ctx, character, items[0].Code, fetched)
	if err != nil {
		return err
	}
	slog.Debug("skill response", "response", skillresp)

	if resp.StatusCode() != 200 {
		return errors.New(resp.Status())
	}

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
