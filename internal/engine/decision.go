package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Operation is a type of event we want a character to do
// ideally this is an event that is run until a stop value is returned
type Operation func(ctx context.Context, r *actions.Runner, character models.Character) bool

// BuildInventory commands a character to focus on building their inventory
// for harvestable items
func BuildInventory(ctx context.Context, r *actions.Runner, character string) error {
	operations := []Operation{gather, bank}
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("get character info: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	currentIndex := 1
	for {
		select {
		case <-ctx.Done():
			slog.Debug("operation loop canceled.")
			return nil
		default:
			currentIndex = (currentIndex + 1) % len(operations)
			for !operations[currentIndex](ctx, r, c) {
				select {
				case <-ctx.Done():
					slog.Debug("engine canceled during processing.")
					return nil
				default:
					slog.Debug("running operations")
				}
			}
		}
	}
}

// Operation loops

func gather(ctx context.Context, r *actions.Runner, character models.Character) bool {
	for {
		select {
		case <-ctx.Done():
			slog.Debug("gather context closed")
			return true
		default:
			slog.Debug("gathering")
			err := Gather(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			slog.Debug("gathering done")
			return true
		}
	}
}

func bank(ctx context.Context, r *actions.Runner, character models.Character) bool {
	for {
		select {
		case <-ctx.Done():
			slog.Debug("banking context closed")
			return true
		default:
			slog.Debug("banking")
			err := DepositAll(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			slog.Debug("banking done")
			return true
		}
	}
}