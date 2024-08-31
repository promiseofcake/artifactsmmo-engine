package engine

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Move physically moves a character if they aren't already there
func Move(ctx context.Context, r *actions.Runner, character string, coords models.Coords) error {
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	if c.X != coords.X || c.Y != coords.Y {
		m, mErr := r.Move(ctx, character, coords.X, coords.Y)
		if mErr != nil {
			return fmt.Errorf("failed to move: %w", mErr)
		}
		cooldown := time.Until(m.CooldownSchema.Expiration)
		slog.Debug("moved to location", "coords", coords, "char", character, "cooldown", cooldown)
		c.CharacterSchema = m.CharacterResponse.CharacterSchema
		time.Sleep(cooldown)
	} else {
		slog.Debug("character already at location, skipping", "coords", coords, "char", character)
	}

	return nil
}

// Travel facilitates travel to the nearest location for a given type/code
func Travel(ctx context.Context, r *actions.Runner, character string, location models.Location) error {
	l := slog.With("character", character)
	maps, err := r.GetMaps(ctx, client.GetAllMapsMapsGetParamsContentType(location.Type))
	if err != nil {
		l.Error("failed to get maps", "error", err)
		return err
	}

	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	coords := models.Coords{}

	distance := math.MaxInt
	for _, m := range maps {
		if m.Code == location.Code {
			temp := models.CalculateDistance(m.Coords, models.Coords{X: c.X, Y: c.Y})
			if temp < distance {
				distance = temp
				coords.X = m.Coords.X
				coords.Y = m.Coords.Y
			}
		}
	}
	l.Debug("location found", "type", location.Type, "code", location.Code, "coords", coords)

	err = Move(ctx, r, character, coords)
	if err != nil {
		l.Error("failed to move", "error", err)
		return err
	}

	return nil
}
