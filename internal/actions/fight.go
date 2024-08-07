package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Fight attacks the mob at the current position for the given character
func (r *Runner) Fight(ctx context.Context, character string) (*FightResponse, error) {
	resp, err := r.Client.ActionFightMyNameActionFightPostWithResponse(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("failed to fight: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fight: %s (%d)", resp.Body, resp.StatusCode())
	}

	return &FightResponse{
		FightResponse: resp.JSON200.Data.Fight,
		Response: Response{
			CharacterResponse: models.Character{CharacterSchema: resp.JSON200.Data.Character},
			CooldownSchema:    resp.JSON200.Data.Cooldown,
		},
	}, nil

}
