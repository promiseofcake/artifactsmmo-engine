package actions

import (
	"context"
	"fmt"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Deposit deposits an item and quantity into the bank
func (r *Runner) Deposit(ctx context.Context, character string, code string, qty int) (*BankResponse, error) {
	resp, err := r.Client.ActionDepositBankMyNameActionBankDepositPostWithResponse(
		ctx,
		character,
		client.ActionDepositBankMyNameActionBankDepositPostJSONRequestBody{
			Code:     code,
			Quantity: qty,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to deposit: %w", err)
	}
	return &BankResponse{
		Item:      resp.JSON200.Data.Item,
		BankItems: resp.JSON200.Data.Bank,
		Response: Response{
			CharacterResponse: models.Character{CharacterSchema: resp.JSON200.Data.Character},
			CooldownSchema:    resp.JSON200.Data.Cooldown,
		},
	}, nil
}
