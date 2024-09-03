package engine

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

var NoItemsToRefine = errors.New("no items to refine")

func Refine(ctx context.Context, r *actions.Runner, character string) error {
	var err error
	// assign character
	l := logging.Get(ctx)

	defer func() {
		// Note that while correct uses of TryLock do exist, they are rare,
		// and use of TryLock is often a sign of a deeper problem
		// in a particular use of mutexes.
		r.RefineMutex.TryLock()
		r.RefineMutex.Unlock()
	}()

	// start by traveling to the bank to reduce the mutex lock time
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

	// get all bank items, determine what's available to refine
	// but lock here so we don't have contention for items
	l.Debug("waiting for refine lock")
	r.RefineMutex.Lock()
	banked, err := r.GetBankItems(ctx)
	if err != nil {
		l.Error("failed to get bank items", "character", character, "error", err)
		return err
	}

	// given all the banked items, get information on each of them and call
	// them resources
	var resources models.Items
	for _, b := range banked {
		item, iErr := r.GetItem(ctx, b.Code)
		if iErr != nil {
			l.Error("failed to get item info", "character", character, "error", iErr)
			return iErr
		}

		item.Quantity = b.Quantity
		// fishing / cooking is the only thing that is 1:1, we need more math
		if item.Type == "resource" && item.Craft == nil {
			resources = append(resources, item)
		}
	}

	// get character state
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		l.Error("failed to get character", "character", character, "error", err)
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

		minLevel := int(math.Max(0, float64(refineLevel-10)))
		items, iErr := r.GetItems(ctx, minLevel, refineLevel, skillType, res.Code)
		if iErr != nil {
			return iErr
		}
		if len(items) == 0 {
			continue
		}

		for _, item := range items {
			cs, cErr := item.Craft.AsCraftSchema()
			if cErr != nil {
				l.Error("failed to get craft schema", "error", cErr)
			}
			for _, n := range *cs.Items {
				item.RawQuantity = n.Quantity
				item.RawCode = n.Code
				break
			}
			refinable = append(refinable, item)
		}
	}

	if len(refinable) == 0 {
		return NoItemsToRefine
	}

	// given the first item in the list (we should sort it)
	// determine how much we want to withdraw.
	slices.SortFunc(refinable, func(a, b models.Item) int {
		return cmp.Compare(a.Level, b.Level)
	})

	// look over all the resources we have in the bank
	// look over all the refinable items that match that
	// resource, and then go for the first one.
	var resourceToRefine = refinable[0]
	for _, res := range resources {
		if res.Code == resourceToRefine.RawCode {
			resourceToRefine.Quantity = res.Quantity
		}
	}
	l.Info("preparing to refine", "resource", resourceToRefine.Name)

	sets := int(math.Floor(float64(c.InventoryMaxItems) / float64(resourceToRefine.RawQuantity)))
	qty := sets * resourceToRefine.RawQuantity

	l.Info("withdrawing", "resource", resourceToRefine.RawCode, "quantity", qty)
	resp, err := r.Withdraw(ctx, character, resourceToRefine.RawCode, qty)
	if err != nil {
		return err
	}
	r.RefineMutex.Unlock()

	cooldown := time.Until(resp.CooldownSchema.Expiration)
	c.CharacterSchema = resp.CharacterResponse.CharacterSchema
	time.Sleep(cooldown)

	// need to travel to refinement location
	l.Info("traveling to workshop", "skill", resourceToRefine.Skill)
	err = Travel(ctx, r, character, models.Location{
		Code: resourceToRefine.Skill,
		Type: "workshop",
	})
	if err != nil {
		return err
	}

	// need to refine the item
	l.Info("refining", "resource", resourceToRefine.Code, "quantity", sets)
	skillresp, err := r.Craft(ctx, character, resourceToRefine.Code, sets)
	if err != nil {
		return fmt.Errorf("failed to craft %s, %d, code: %w", resourceToRefine.Code, qty, err)
	}
	l.Info("skill response", "response", skillresp.SkillInfo)

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

func RefineAll(ctx context.Context, r *actions.Runner, character string) error {
	l := logging.Get(ctx)
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("get character info: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			l.Debug("refine loop canceled.")
			return nil
		default:
			l.Debug("refining")
			rErr := Refine(ctx, r, c.Name)
			if rErr == nil || errors.Is(rErr, NoItemsToRefine) {
				// no issue if nothing to refine
				return nil
			} else {
				l.Error("failed to refine", "character", character, "error", rErr)
				return rErr
			}
		}
	}
}
