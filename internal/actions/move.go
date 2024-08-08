package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Move changes the x, y position the given character
func (r *Runner) Move(ctx context.Context, character string, x, y int) (*Response, error) {
	resp, err := r.Client.ActionMoveMyNameActionMovePostWithResponse(ctx, character, client.ActionMoveMyNameActionMovePostJSONRequestBody{
		X: x,
		Y: y,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to move: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status failure (%d), message: %s", resp.StatusCode(), resp.Body)
	}

	return &Response{
		CharacterResponse: models.Character{CharacterSchema: resp.JSON200.Data.Character},
		CooldownSchema:    resp.JSON200.Data.Cooldown,
	}, nil
}
