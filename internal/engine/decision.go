package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

type Operation func(ctx context.Context, r *actions.Runner, character models.Character) bool

func BuildInventory(ctx context.Context, r *actions.Runner, character string) error {
	operations := []Operation{gather, bank}
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		return fmt.Errorf("get character info: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	currentIndex := 1
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Operation loop canceled.")
			return nil
		default:
			currentIndex = (currentIndex + 1) % len(operations)
			for !operations[currentIndex](ctx, r, c) {
				select {
				case <-ctx.Done():
					fmt.Println("engine canceled during processing.")
					return nil
				default:
					slog.Info("running operations")
				}
			}
		}
	}
}

func gather(ctx context.Context, r *actions.Runner, character models.Character) bool {
	for {
		select {
		case <-ctx.Done():
			slog.Info("gather context closed")
			return true
		default:
			err := Gather(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			slog.Info("gathering done")
			return true
		}
	}
}

func bank(ctx context.Context, r *actions.Runner, character models.Character) bool {
	for {
		select {
		case <-ctx.Done():
			slog.Info("banking context closed")
			return true
		default:
			slog.Info("banking")
			err := Deposit(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			slog.Info("banking done")
			return true
		}
	}
}
