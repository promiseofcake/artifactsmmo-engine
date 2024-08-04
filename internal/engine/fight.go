package engine

import (
	"context"
	"fmt"
	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log/slog"
	"math"
	"time"
)

func Fight(ctx context.Context, r *actions.Runner, character string) error {
	char, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to get character %w", err)
	}

	monsterLocations, err := r.GetMaps(ctx, client.Monster)
	if err != nil {
		return fmt.Errorf("failed to get monster locations %w", err)
	}

	minLevel := int(math.Round(math.Floor(float64(char.Level) - (float64(char.Level) * float64(.50)))))
	maxLevel := int(math.Round(math.Ceil(float64(char.Level) + (float64(char.Level) * float64(.10)))))
	monsterInfo, err := r.GetMonsters(ctx, minLevel, maxLevel)
	if err != nil {
		return fmt.Errorf("failed to get monster information %w", err)
	}

	loc := actions.LocationsToMap(monsterLocations)
	mon := actions.MonstersToMap(monsterInfo)
	mon.FindMonsters(loc)

	var monster actions.Monster
	// pick a random monster
	for _, m := range mon {
		monster = *m
		break
	}

	resp, err := r.Move(ctx, character, monster.Location.X, monster.Location.Y)
	if err != nil {
		return fmt.Errorf("failed to move to monster %w", err)
	}
	cooldown := time.Until(resp.CooldownSchema.Expiration)
	slog.Info("moved to monster", "char", character, "cooldown", cooldown)
	char = &resp.CharacterResponse
	time.Sleep(cooldown)

	for {
		f, err := r.Fight(ctx, character)
		if err != nil {
			return fmt.Errorf("failed to fight monster %w", err)
		}
		cooldown := time.Until(f.CooldownSchema.Expiration)
		slog.Info("fight results",
			"results", f.FightResponse,
			"cooldown", cooldown,
		)
		char = &f.CharacterResponse
		time.Sleep(cooldown)
	}
}
