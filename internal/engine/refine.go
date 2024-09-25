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
	// lock here, so we don't have contention for items
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
			resources = append(resources, &item)
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
	// and put these into a new list
	var refinable models.Items
	for _, res := range resources {
		var refineLevel int
		var skillType string

		// given the resource's subtype, lookup this characters
		// skill level, and fetch all items one could make
		// with the current given skill level
		switch res.Subtype {
		case string(client.ResourceSchemaSkillWoodcutting):
			refineLevel = c.WoodcuttingLevel
			skillType = string(client.CraftSchemaSkillWoodcutting)
		case string(client.ResourceSchemaSkillMining):
			refineLevel = c.MiningLevel
			skillType = string(client.CraftSchemaSkillMining)
		case string(client.ResourceSchemaSkillFishing):
			refineLevel = c.CookingLevel
			skillType = string(client.CraftSchemaSkillCooking)
		}

		// min level current level - 10
		// TODO make overridable
		//minLevel := int(math.Max(0, float64(refineLevel-10)))

		// allow all items to be refined
		minLevel := 0

		// get items that match
		items, iErr := r.GetItems(ctx, minLevel, refineLevel, skillType, res.Code)
		if iErr != nil {
			return iErr
		}
		if len(items) == 0 {
			continue
		}
		refinable = append(refinable, items...)
	}

	if len(refinable) == 0 {
		return NoItemsToRefine
	}

	// sort list with higher level items first
	slices.SortFunc(refinable, func(a, b *models.Item) int {
		return cmp.Compare(b.Level, a.Level)
	})

	// look over all the resources we have in the bank
	// look over all the refinable items that match that
	// resource, and then go for the first one.

	// bad loop
	var totalResourceCount int
	var available models.Items
	for _, resourceToRefine := range refinable {
		totalResourceCount = 0
		for _, mat := range resourceToRefine.CraftMaterials {
			totalResourceCount += mat.CostPerResource
			for _, res := range resources {
				if mat.RequiredCode == res.Code {
					mat.Available = res.Quantity
				}
			}
		}

		maxSetsByInventory := int(math.Floor(float64(c.InventoryMaxItems) / float64(totalResourceCount)))
		setCount := maxSetsByInventory

		for _, mat := range resourceToRefine.CraftMaterials {
			maxSetsByResource := mat.Available / mat.CostPerResource
			if maxSetsByResource < setCount {
				setCount = maxSetsByResource
			}
		}

		if setCount > 0 {
			resourceToRefine.Quantity = setCount
			available = append(available, resourceToRefine)
		}
	}

	resourceToRefine := available[0]
	for _, mat := range resourceToRefine.CraftMaterials {
		qty := resourceToRefine.Quantity * mat.CostPerResource
		l.Info("withdrawing item", "code", mat.RequiredCode, "qty", qty)
		resp, wErr := r.Withdraw(ctx, character, mat.RequiredCode, qty)
		if wErr != nil {
			return wErr
		}
		cooldown := time.Until(resp.CooldownSchema.Expiration)
		c.CharacterSchema = resp.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}
	r.RefineMutex.Unlock()

	l.Info("preparing to refine", "resource", resourceToRefine.Name, "qty", resourceToRefine.Quantity)

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
	skillresp, err := r.Craft(ctx, character, resourceToRefine.Code, resourceToRefine.Quantity)
	if err != nil {
		return fmt.Errorf("failed to craft %s, %d, code: %w", resourceToRefine.Code, resourceToRefine.Quantity, err)
	}
	l.Info("skill response", "response", skillresp.SkillInfo)

	cooldown := time.Until(skillresp.Response.CooldownSchema.Expiration)
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
