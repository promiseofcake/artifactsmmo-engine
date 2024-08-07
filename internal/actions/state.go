package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

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

// GetMaps fetches world state based upon a given content type
func (r *Runner) GetMaps(ctx context.Context, contentType client.GetAllMapsMapsGetParamsContentType) (models.Locations, error) {
	resp, err := r.Client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
		ContentType: &contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maps for content: %s %w", contentType, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch maps for content: %s (%d)", resp.Body, resp.StatusCode())
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
		return nil, fmt.Errorf("failed to fetch monsters: %s (%d)", resp.Body, resp.StatusCode())
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
