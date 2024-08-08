package engine

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Gather will attempt to Gather resources until the character should bank
func Gather(ctx context.Context, r *actions.Runner, character string) error {
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Error("failed to get character", "error", err)
		return err
	}

	resourceLoations, err := r.GetMaps(ctx, client.Resource)
	if err != nil {
		slog.Error("failed to get resource locations", "error", err)
		return err
	}

	skill := c.ChooseWeakestSkill()
	resourceInfo, err := r.GetResources(ctx, skill.Code, skill.MinLevel, skill.CurrentLevel)
	if err != nil {
		slog.Error("failed to get resources", "error", err)
		return err
	}

	loc := models.LocationsToMap(resourceLoations)
	res := models.ResourcesToMap(resourceInfo)
	// TODO there are more than one resource available, we should move to the one closet to the bank
	// implement with manhattan distance
	res.FindResources(loc)

	resources := res.ToSlice()
	slices.SortFunc(resources, func(a, b models.Resource) int {
		return cmp.Compare(b.Level, a.Level)
	})

	if len(resources) == 0 {
		err = fmt.Errorf("no suitable resources found")
		slog.Error("failed to gather", "error", err)
		return err
	}

	// begin
	resource := resources[0]
	slog.Debug("choosing to gather", "resource", resource)

	// check if we should bank straight away
	if c.ShouldBank() {
		slog.Debug("character will bank")
		return nil
	}

	// go to resource
	if c.X != resource.GetCoords().X && c.Y != resource.GetCoords().Y {
		m, mErr := r.Move(ctx, c.Name, resource.GetCoords().X, resource.GetCoords().Y)
		if mErr != nil {
			slog.Error("failed to move", "error", err)
			return err
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Debug("moved to resource", "resource", resource, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}

	// harvest resource until we should stop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			g, gErr := r.Gather(ctx, character)
			if gErr != nil {
				slog.Error("failed to gather", "error", gErr)
				return gErr
			}
			cooldown := time.Until(g.CooldownSchema.Expiration)
			slog.Debug("gathered resource", "resource", resource, "result", g.SkillInfo, "cooldown", cooldown)
			c.CharacterSchema = g.CharacterResponse.CharacterSchema
			time.Sleep(cooldown)

			if c.ShouldBank() {
				slog.Debug("character will bank")
				return nil
			}
		}
	}
}
