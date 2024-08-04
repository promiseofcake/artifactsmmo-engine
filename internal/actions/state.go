package actions

import (
	"context"
	"fmt"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log/slog"
	"net/http"
)

// GetMyCharacterInfo returns current info and status about your own specific character
func (r *Runner) GetMyCharacterInfo(ctx context.Context, character string) (*CharacterResponse, error) {
	resp, err := r.Client.GetMyCharactersMyCharactersGetWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch characters: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch characters: %s (%d)", resp.Body, resp.StatusCode())
	}

	for _, c := range resp.JSON200.Data {
		if c.Name == character {
			return &CharacterResponse{
				CharacterSchema: c,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to find character: %s", character)
}

// GetMaps fetches world state based upon a given content type
func (r *Runner) GetMaps(ctx context.Context, contentType client.GetAllMapsMapsGetParamsContentType) (Locations, error) {
	resp, err := r.Client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
		ContentType: &contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maps for content: %s %w", contentType, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch maps for content: %s (%d)", resp.Body, resp.StatusCode())
	}

	var mc Locations
	for _, m := range resp.JSON200.Data {
		s, err := m.Content.AsMapContentSchema()
		if err != nil {
			slog.Error("failed to extract map content schema", "error", err)
		}

		c := Location{
			Name: m.Name,
			Skin: m.Skin,
			X:    m.X,
			Y:    m.Y,
			Code: s.Code,
			Type: s.Type,
		}

		mc = append(mc, c)
	}
	return mc, nil
}

// GetMonsters fetches monster world state based upon a given content type
func (r *Runner) GetMonsters(ctx context.Context, min, max int) (Monsters, error) {
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

	var mm Monsters
	for _, m := range resp.JSON200.Data {
		monster := Monster{
			Name:     m.Name,
			Code:     m.Code,
			Level:    m.Level,
			Location: Location{},
		}
		mm = append(mm, monster)
	}

	return mm, nil
}
