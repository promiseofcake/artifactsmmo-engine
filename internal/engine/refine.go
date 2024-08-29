package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/actions"
	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

func Refine(ctx context.Context, r *actions.Runner, character string) error {
	var err error
	// need to determine what is available to refine
	banked, err := r.GetBankItems(ctx)
	if err != nil {
		slog.Debug("failed to get bank items", "error", err)
		return err
	}

	// lookup all the bank items and make a list of reosurces
	var resources models.Items
	for _, b := range banked {
		item, iErr := r.GetItem(ctx, b.Code)
		if iErr != nil {
			slog.Debug("failed to get item info", "error", iErr)
			return iErr
		}

		// only look for resources we can work with
		item.Quantity = b.Quantity
		if item.Type == "resource" {
			resources = append(resources, item)
		}
	}

	fmt.Printf("wait")

	// given a list of all the current resources and their quantity, determine what to refine
	// subtype is the thing itself
	c, err := r.GetMyCharacterInfo(ctx, character)
	if err != nil {
		slog.Debug("failed to get character", "error", err)
		return err
	}

	// what to refine?
	for _, have := range resources {
		// given subtype, lookup level
		var cl int
		switch have.Subtype {
		case string(client.Woodcutting):
			cl = c.WoodcuttingLevel
		case string(client.Mining):
			cl = c.MiningLevel
		case string(client.Fishing):
			cl = c.FishingLevel
		}

		ii, err := r.GetItems(ctx, 0, cl, have.Subtype, have.Code)
		if err != nil {
			return err
		}
		fmt.Printf("%+v", ii)
	}

	// need to travel to bank

	// need to withdraw

	// need to travel to refinement location

	// need to refine

	// need to return to bank and deposit
	err = DepositAll(ctx, r, character)
	if err != nil {
		return err
	}

	return errors.New("refine not implemented yet")
}
