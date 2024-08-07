package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Craft crafts the given item, with the given quantity and assumes the character is in the correct
// map position
func (r *Runner) Craft(ctx context.Context, character string, code string, quantity int) (*SkillResponse, error) {
	req := client.ActionCraftingMyNameActionCraftingPostJSONRequestBody{
		Code:     code,
		Quantity: &quantity,
	}

	resp, err := r.Client.ActionCraftingMyNameActionCraftingPostWithResponse(ctx, character, req)
	if err != nil {
		return nil, fmt.Errorf("failed to craft %s (%d): %w", code, quantity, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to craft: %s (%d)", resp.Body, resp.StatusCode())
	}

	return &SkillResponse{
		SkillInfo: resp.JSON200.Data.Details,
		Response: Response{
			CharacterResponse: models.Character{CharacterSchema: resp.JSON200.Data.Character},
			CooldownSchema:    resp.JSON200.Data.Cooldown,
		},
	}, nil
}
