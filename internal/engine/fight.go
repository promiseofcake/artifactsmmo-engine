package engine

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"

	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Fight will attempt to find and fight appropriate monsters
func Fight(ctx context.Context, r *actions.Runner, character string) error {
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Error("failed to get character", "error", err)
		return err
	}

	monsterLocations, err := r.GetMaps(ctx, client.Monster)
	if err != nil {
		slog.Error("failed to get monster locations", "error", err)
		return err
	}

	minLevel := int(math.Round(math.Floor(float64(c.Level) - (float64(c.Level) * float64(.50)))))
	maxLevel := int(math.Round(math.Ceil(float64(c.Level) + (float64(c.Level) * float64(.10)))))
	monsterInfo, err := r.GetMonsters(ctx, minLevel, maxLevel)
	if err != nil {
		slog.Error("failed to get monsters", "error", err)
		return err
	}

	loc := models.LocationsToMap(monsterLocations)
	mon := models.MonstersToMap(monsterInfo)
	mon.FindMonsters(loc)

	var monster models.Monster
	// pick a random monster
	for _, m := range mon {
		monster = *m
		break
	}

	err = Move(ctx, r, character, monster.GetCoords())
	if err != nil {
		slog.Error("failed to move to monster", "error", err)
		return err
	}

	for {
		f, fErr := r.Fight(ctx, character)
		if fErr != nil {
			slog.Error("failed to fight monster", "error", fErr)
			return fErr
		}
		fCooldown := time.Until(f.CooldownSchema.Expiration)
		slog.Debug("fight results",
			"results", f.FightResponse,
			"cooldown", fCooldown,
		)
		c = f.CharacterResponse
		time.Sleep(fCooldown)
	}
}
