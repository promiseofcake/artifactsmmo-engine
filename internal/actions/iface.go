package actions

import (
	"context"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

type Runner interface {
	Deposit(ctx context.Context, character string, code string, qty int) (*BankResponse, error)
	Withdraw(ctx context.Context, character string, code string, qty int) (*BankResponse, error)
	Craft(ctx context.Context, character string, code string, quantity int) (*SkillResponse, error)
	Fight(ctx context.Context, character string) (*FightResponse, error)
	Gather(ctx context.Context, character string) (*SkillResponse, error)
	Move(ctx context.Context, character string, x, y int) (*Response, error)

	GetBankItems(ctx context.Context) (models.SimpleItems, error)
	GetMyCharacterInfo(ctx context.Context, character string) (models.Character, error)
	GetMapsByContentCode(ctx context.Context, contentCode string) (models.Locations, error)
	GetMapsByContentType(ctx context.Context, contentType client.GetAllMapsMapsGetParamsContentType) (models.Locations, error)
	GetItem(ctx context.Context, code string) (models.Item, error)
	GetItems(ctx context.Context, min, max int, skill string, material string) (models.Items, error)
	GetMonsters(ctx context.Context, min, max int) (models.Monsters, error)
	GetResourcesByDrop(ctx context.Context, drop string) (models.Resources, error)
	GetResourcesBySkill(ctx context.Context, skill client.ResourceSchemaSkill, min, max int) (models.Resources, error)
}
