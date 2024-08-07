package actions

import (
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Response is a generic return value for most actions
// it includes the most common return values for Character State
// and request cooldown
type Response struct {
	CharacterResponse models.Character
	CooldownSchema    client.CooldownSchema
}

// GetCooldownDuration returns the duration for the Response cooldown
func (g *Response) GetCooldownDuration() time.Duration {
	return time.Until(g.CooldownSchema.Expiration)
}

// FightResponse wraps a generic Response with Fight related data
type FightResponse struct {
	Response
	FightResponse client.FightSchema
}

// SkillResponse wraps a generic Response with Skill related data
// Used for craft, gather
type SkillResponse struct {
	Response
	SkillInfo client.SkillInfoSchema
}

// BankResponse wraps a generic Response with Banking related data
type BankResponse struct {
	Response
	BankItems []client.SimpleItemSchema
	Item      client.ItemSchema
}
