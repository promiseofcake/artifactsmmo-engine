package engine

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Gather will attempt to Gather resources until the character should bank
func Gather(ctx context.Context, r *actions.Runner, character string) error {
	type Resource struct {
		Code   string
		Coords models.Coords
	}

	// hardcoded resources
	resources := []Resource{
		{
			Code: "ash_tree",
			Coords: models.Coords{
				X: -1,
				Y: 0,
			},
		},
		{
			Code: "copper_rocks",
			Coords: models.Coords{
				X: 2,
				Y: 0,
			},
		},
	}
	resource := resources[rand.Intn(2)]
	slog.Debug("choosing to gather", "resource", resource)

	// get all character info
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Error("failed to get character", "error", err)
		return err
	}

	// check if we should bank straight away
	if c.ShouldBank() {
		slog.Debug("character will bank")
		return nil
	}

	// go to resource
	if c.X != resource.Coords.X && c.Y != resource.Coords.Y {
		m, mErr := r.Move(ctx, c.Name, resource.Coords.X, resource.Coords.Y)
		if mErr != nil {
			slog.Error("failed to move", "error", err)
			return err
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Debug("moved to resource", "resource", resource, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}

	// mine resource until we should stop
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
