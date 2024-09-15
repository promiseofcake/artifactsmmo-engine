package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Gather performs resource gathering at the current position for the given character
func (r *RealRunner) Gather(ctx context.Context, character string) (*SkillResponse, error) {
	resp, err := r.Client.ActionGatheringMyNameActionGatheringPostWithResponse(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("failed to gather: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status failure (%d), message: %s", resp.StatusCode(), resp.Body)
	}

	return &SkillResponse{
		SkillInfo: resp.JSON200.Data.Details,
		Response: Response{
			CharacterResponse: models.Character{CharacterSchema: resp.JSON200.Data.Character},
			CooldownSchema:    resp.JSON200.Data.Cooldown,
		},
	}, nil
}
