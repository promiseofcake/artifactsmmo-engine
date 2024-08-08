package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Deposit deposits an item and quantity into the bank
func (r *Runner) Deposit(ctx context.Context, character string, code string, qty int) (*BankResponse, error) {
	body := client.ActionDepositBankMyNameActionBankDepositPostJSONRequestBody{
		Code:     code,
		Quantity: qty,
	}

	resp, err := r.Client.ActionDepositBankMyNameActionBankDepositPostWithResponse(
		ctx,
		character,
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to deposit: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status failure (%d), message: %s", resp.StatusCode(), resp.Body)
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
