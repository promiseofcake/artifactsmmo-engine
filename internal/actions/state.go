package actions

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// GetBankItems returns all items in the bank
func (r *Runner) GetBankItems(ctx context.Context) (models.SimpleItems, error) {
	resp, err := r.Client.GetBankItemsMyBankItemsGetWithResponse(ctx, &client.GetBankItemsMyBankItemsGetParams{})
	if err != nil {
		return models.SimpleItems{}, fmt.Errorf("failed to get bank items: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return models.SimpleItems{}, fmt.Errorf("failed to get bank items: %s (%d)", resp.Body, resp.StatusCode())
	}

	var bank models.SimpleItems
	for _, i := range resp.JSON200.Data {
		item := models.SimpleItem{
			Code:     i.Code,
			Quantity: i.Quantity,
		}
		bank = append(bank, item)
	}

	return bank, nil
}

// GetMyCharacterInfo returns current info and status about your own specific character
func (r *Runner) GetMyCharacterInfo(ctx context.Context, character string) (models.Character, error) {
	resp, err := r.Client.GetMyCharactersMyCharactersGetWithResponse(ctx)
	if err != nil {
		return models.Character{}, fmt.Errorf("failed to get character info: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return models.Character{}, fmt.Errorf("failed to get character info: %s (%d)", resp.Body, resp.StatusCode())
	}

	for _, c := range resp.JSON200.Data {
		if c.Name == character {
			return models.Character{
				CharacterSchema: c,
			}, nil
		}
	}

	return models.Character{}, fmt.Errorf("failed to find character: %s", character)
}

// GetMapsByContentCode fetches world state based upon a given content code
func (r *Runner) GetMapsByContentCode(ctx context.Context, contentCode string) (models.Locations, error) {
	resp, err := r.Client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
		ContentCode: &contentCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maps for content: %s %w", contentCode, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch maps: %s (%d)", resp.Body, resp.StatusCode())
	}

	var locs models.Locations
	for _, l := range resp.JSON200.Data {
		s, dataErr := l.Content.AsMapContentSchema()
		if dataErr != nil {
			return nil, fmt.Errorf("failed to extra map content schema: %w", err)
		}

		loc := models.Location{
			Name: l.Name,
			Skin: l.Skin,
			Coords: models.Coords{
				X: l.X,
				Y: l.Y,
			},
			Code: s.Code,
			Type: s.Type,
		}

		locs = append(locs, loc)
	}
	return locs, nil
}

// GetMapsByContentType fetches world state based upon a given content type
func (r *Runner) GetMapsByContentType(ctx context.Context, contentType client.GetAllMapsMapsGetParamsContentType) (models.Locations, error) {
	resp, err := r.Client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
		ContentType: &contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maps for content: %s %w", contentType, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch maps: %s (%d)", resp.Body, resp.StatusCode())
	}

	var locs models.Locations
	for _, l := range resp.JSON200.Data {
		s, dataErr := l.Content.AsMapContentSchema()
		if dataErr != nil {
			return nil, fmt.Errorf("failed to extra map content schema: %w", err)
		}

		loc := models.Location{
			Name: l.Name,
			Skin: l.Skin,
			Coords: models.Coords{
				X: l.X,
				Y: l.Y,
			},
			Code: s.Code,
			Type: s.Type,
		}

		locs = append(locs, loc)
	}
	return locs, nil
}

// GetItem returns information about an item
func (r *Runner) GetItem(ctx context.Context, code string) (models.Item, error) {
	resp, err := r.Client.GetItemItemsCodeGetWithResponse(ctx, code)
	if err != nil {
		return models.Item{}, fmt.Errorf("failed to get item with code: %s %w", code, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return models.Item{}, fmt.Errorf("failed to get item: %s (%d)", resp.Body, resp.StatusCode())
	}

	return models.Item{ItemSchema: resp.JSON200.Data.Item}, nil
}

// GetItems searches for an item
func (r *Runner) GetItems(ctx context.Context, min, max int, skill string, material string) (models.Items, error) {
	s := client.GetAllItemsItemsGetParamsCraftSkill(skill)

	resp, err := r.Client.GetAllItemsItemsGetWithResponse(ctx, &client.GetAllItemsItemsGetParams{
		CraftSkill:    &s,
		CraftMaterial: &material,
		MinLevel:      &min,
		MaxLevel:      &max,
	})
	if err != nil {
		return models.Items{}, err
	}

	if resp.StatusCode() != http.StatusOK {
		return models.Items{}, err
	}

	var items models.Items
	for _, i := range resp.JSON200.Data {
		a := models.Item{ItemSchema: i}

		cs, cErr := a.Craft.AsCraftSchema()
		if cErr != nil {
			return models.Items{}, fmt.Errorf("failed to get craft schema for: %s, error: %w", i.Code, cErr)
		}

		var inputs []*models.CraftResource

		a.Skill = string(*cs.Skill)
		required := *cs.Items
		for _, ii := range required {
			inputs = append(inputs, &models.CraftResource{RequiredCode: ii.Code, CostPerResource: ii.Quantity})
		}
		a.CraftMaterials = inputs

		items = append(items, &a)
	}

	return items, nil
}

// GetMonsters fetches monster world state based upon a given content type
func (r *Runner) GetMonsters(ctx context.Context, min, max int) (models.Monsters, error) {
	resp, err := r.Client.GetAllMonstersMonstersGetWithResponse(ctx, &client.GetAllMonstersMonstersGetParams{
		MinLevel: &min,
		MaxLevel: &max,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monsters for levels: %d-%d %w", min, max, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status failure (%d), message: %s", resp.StatusCode(), resp.Body)
	}

	var monsters models.Monsters
	for _, m := range resp.JSON200.Data {
		monster := models.Monster{
			Name:     m.Name,
			Code:     m.Code,
			Level:    m.Level,
			Location: models.Location{},
		}
		monsters = append(monsters, monster)
	}

	return monsters, nil
}

func (r *Runner) GetResourcesByDrop(ctx context.Context, drop string) (models.Resources, error) {
	resp, err := r.Client.GetAllResourcesResourcesGetWithResponse(ctx, &client.GetAllResourcesResourcesGetParams{
		Drop: &drop,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources for drop %s, %w", drop, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status failure (%d), message: %s", resp.StatusCode(), resp.Body)
	}

	var resources models.Resources
	for _, res := range resp.JSON200.Data {

		locations, lErr := r.GetMapsByContentCode(ctx, res.Code)
		if lErr != nil || len(locations) == 0 {
			return nil, fmt.Errorf("failed to find resource locations: %w", err)
		}

		resource := models.Resource{
			Name:     res.Name,
			Code:     res.Code,
			Skill:    res.Skill,
			Level:    res.Level,
			Location: locations[0], // todo allow more locations
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// GetResourcesBySkill returns all resources (and location) for resources in a given skill / level range
func (r *Runner) GetResourcesBySkill(ctx context.Context, skill client.ResourceSchemaSkill, min, max int) (models.Resources, error) {
	if min < 0 {
		min = 0
	}

	if max < 0 {
		max = 0
	}

	s := client.GetAllResourcesResourcesGetParamsSkill(skill)

	resp, err := r.Client.GetAllResourcesResourcesGetWithResponse(ctx, &client.GetAllResourcesResourcesGetParams{
		MinLevel: &min,
		MaxLevel: &max,
		Skill:    &s,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources for skill %s, levels: %d-%d %w", skill, min, max, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status failure (%d), message: %s", resp.StatusCode(), resp.Body)
	}

	var resources models.Resources
	for _, res := range resp.JSON200.Data {
		locations, lErr := r.GetMapsByContentCode(ctx, res.Code)
		if lErr != nil {
			return nil, fmt.Errorf("failed to find resource locations: %w", err)
		}
		if len(locations) == 0 {
			logging.Get(ctx).Info("skipping resource locations: no locations found", "resource", res)
			continue
		}

		resource := models.Resource{
			Name:     res.Name,
			Code:     res.Code,
			Skill:    res.Skill,
			Level:    res.Level,
			Location: locations[0], // todo allow more locations
		}
		resources = append(resources, resource)
	}

	slices.SortFunc(resources, func(a, b models.Resource) int {
		return cmp.Compare(b.Level, a.Level)
	})

	return resources, nil
}
