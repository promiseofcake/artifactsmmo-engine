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

func (r *Runner) GetMaps(ctx context.Context, contentType client.GetAllMapsMapsGetParamsContentType) (MapContent, error) {
	resp, err := r.Client.GetAllMapsMapsGetWithResponse(ctx, &client.GetAllMapsMapsGetParams{
		ContentType: &contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maps for content: %s %w", contentType, err)
	}

	var mc MapContent
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
