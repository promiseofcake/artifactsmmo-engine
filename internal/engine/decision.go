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

// Execute commands a character to focus on building their inventory
// for harvestable items
func Execute(ctx context.Context, r *actions.Runner, character string, actions []string, orders models.SimpleItems) error {
	l := logging.Get(ctx)

	var operations []Operation
	for _, op := range actions {
		switch op {
		case "gather":
			// fallthrough
		case "forage":
			operations = append(operations, forage)
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

			l.Debug("checking for orders to fulfil", "orders", orders)
			for _, o := range orders {
				if ShouldFulfilOrder(ctx, r, o) {
					l.Debug("order required to be fulfilled", "order", o)
					oErr := FulfilOrder(ctx, r, character, o)
					if oErr != nil {
						l.Error("failed to fulfil order", "order", o, "error", oErr)
					}
				}
			}

			l.Debug("performing designated tasks", "tasks", operations)
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
func forage(ctx context.Context, r *actions.Runner, character models.Character) bool {
	l := logging.Get(ctx)
	for {
		select {
		case <-ctx.Done():
			l.Debug("foraging context closed")
			return true
		default:
			l.Debug("foraging")
			err := Forage(ctx, r, character.Name)
			if err != nil {
				panic(err)
			}
			l.Debug("foraging done")
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
