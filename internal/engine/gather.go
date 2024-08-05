package engine

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/player"
)

func Gather(ctx context.Context, r *actions.Runner, character string) error {
	type Resource struct {
		Code string
		X    int
		Y    int
	}

	// resources
	resources := []Resource{
		{
			Code: "ash_tree",
			X:    -1,
			Y:    0,
		},
		{
			Code: "copper_rocks",
			X:    2,
			Y:    0,
		},
	}
	resource := resources[rand.Intn(1)]
	slog.Info("we are gathering", "resource", resource)

	// get all character info
	char, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character %w", err)
	}

	// our mutated state
	c := player.Character{
		CharacterSchema: char.CharacterSchema,
	}

	// check if we should bank straight away
	if c.ShouldBank() {
		return nil
	}

	// go to resource
	if c.X != resource.X && c.Y != resource.Y {
		m, err := r.Move(ctx, c.Name, resource.X, resource.Y)
		if err != nil {
			return fmt.Errorf("failed to move %w", err)
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Info("moved to resource", "resource", resource, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	}

	// mine resource until we should stop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			g, err := r.Gather(ctx, character)
			if err != nil {
				return err
			}
			cooldown := time.Until(g.CooldownSchema.Expiration)
			slog.Info("gathered resource", "resource", resource, "result", g.SkillInfo, "cooldown", cooldown)
			c.CharacterSchema = g.CharacterResponse.CharacterSchema

			if c.ShouldBank() {
				slog.Info("time to bank, too many resources")
				return nil
			}
		}
	}
}
