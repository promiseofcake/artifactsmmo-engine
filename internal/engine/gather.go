package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Forage will attempt to Forage resources until the character should bank
func Forage(ctx context.Context, r *actions.Runner, character string) error {
	l := logging.Get(ctx)
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		l.Error("failed to get character", "error", err)
		return err
	}

	skill := c.ChooseWeakestSkill()
	resources, err := r.GetResourcesBySkill(ctx, skill.Code, skill.MinLevel, skill.CurrentLevel)
	if err != nil {
		l.Error("failed to get resources", "error", err)
		return err
	}

	if len(resources) == 0 {
		err = fmt.Errorf("no suitable resources found")
		l.Error("failed to gather", "error", err)
		return err
	}

	// begin
	resource := resources[0]
	l.Debug("choosing to gather", "resource", resource)

	// check if we should bank straight away
	if c.ShouldBank() {
		l.Debug("character will bank")
		return DepositAll(ctx, r, character)
	}

	return Gather(ctx, r, character, resource)
}

// Gather will move to, and gather loop a resource
func Gather(ctx context.Context, r *actions.Runner, character string, resource models.Resource) error {
	l := logging.Get(ctx)

	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		l.Error("failed to get character", "error", err)
		return err
	}

	mErr := Move(ctx, r, character, resource.GetCoords())
	if mErr != nil {
		l.Error("failed to move", "error", mErr)
		return mErr
	}

	// harvest resource until we should stop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			g, gErr := r.Gather(ctx, character)
			if gErr != nil {
				l.Error("failed to gather", "error", gErr)
				return gErr
			}
			cooldown := time.Until(g.CooldownSchema.Expiration)
			l.Info("gathered resource", "resource", resource, "result", g.SkillInfo, "cooldown", cooldown)
			c.CharacterSchema = g.CharacterResponse.CharacterSchema
			time.Sleep(cooldown)

			if c.ShouldBank() {
				l.Debug("character will bank")
				return DepositAll(ctx, r, character)
			}
		}
	}
}
