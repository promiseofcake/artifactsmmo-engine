package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/logging"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Operation is a type of event we want a character to do
// ideally this is an event that is run until a stop value is returned
type Operation func(ctx context.Context, r *actions.Runner, character models.Character) bool

// BuildInventory commands a character to focus on building their inventory
// for harvestable items
func BuildInventory(ctx context.Context, r *actions.Runner, character string, actions []string) error {
	l := logging.Get(ctx)

	var operations []Operation
	for _, op := range actions {
		switch op {
		case "bank":
			operations = append(operations, bank)
		case "gather":
			operations = append(operations, gather)
		case "refine":
			operations = append(operations, refine)
		}
	}

	if len(operations) == 0 {
		slog.Error("nothing to do for character")
		return errors.New("nothing to do for character")
	}

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
			l.Debug("operation loop canceled.")
			return nil
		default:
			// TODO this isnt' right
			currentIndex = (currentIndex + 1) % len(operations)
			for !operations[currentIndex](ctx, r, c) {
				select {
				case <-ctx.Done():
					l.Debug("engine canceled during processing.")
					return nil
				default:
					l.Debug("running operations")
				}
			}
		}
	}
}

// Operation loops
func bank(ctx context.Context, r *actions.Runner, character models.Character) bool {
	l := logging.Get(ctx)
	for {
		select {
		case <-ctx.Done():
			l.Debug("banking context closed")
			return true
		default:
			l.Debug("banking")
			err := DepositAll(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			l.Debug("banking done")
			return true
		}
	}
}

func gather(ctx context.Context, r *actions.Runner, character models.Character) bool {
	l := logging.Get(ctx)
	for {
		select {
		case <-ctx.Done():
			l.Debug("gather context closed")
			return true
		default:
			l.Debug("gathering")
			err := Gather(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			l.Debug("gathering done")
			return true
		}
	}
}

func refine(ctx context.Context, r *actions.Runner, character models.Character) bool {
	l := logging.Get(ctx)
	for {
		select {
		case <-ctx.Done():
			l.Debug("refine context closed")
			return true
		default:
			l.Debug("refining")
			err := RefineAll(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			l.Debug("refining done")
			return true
		}
	}
}
