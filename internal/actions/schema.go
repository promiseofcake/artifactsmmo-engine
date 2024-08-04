package actions

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
)

// Coords defines X, Y coords as signed integers
type Coords struct {
	X int
	Y int
}

// Response is a generic return value for most actions
type Response struct {
	CharacterResponse CharacterResponse
	CooldownSchema    client.CooldownSchema
}

// GetRemainingCooldown returns the remaining cooldown based upon the last action
// this is useful for wait loops
func (g *Response) GetRemainingCooldown() int64 {
	return int64(g.CooldownSchema.RemainingSeconds)
}

// CharacterResponse is a specific wrapper around generic return info
type CharacterResponse struct {
	client.CharacterSchema
}

// GetPosition returns the given Coords for the Character in question
func (c *CharacterResponse) GetPosition() Coords {
	return Coords{
		X: c.X,
		Y: c.Y,
	}
}

// FightResponse wraps a generic Response with Fight related data
type FightResponse struct {
	Response
	FightResponse client.FightSchema
}

// SkillResponse wraps a generic Response with Skill related data
type SkillResponse struct {
	Response
	SkillInfo client.SkillInfoSchema
}

type MapContent []Location

type Location struct {
	Name string `json:"name"`
	Skin string `json:"skin"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Code string `json:"code"`
	Type string `json:"type"`
}
