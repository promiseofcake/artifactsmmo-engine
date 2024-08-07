package actions

import (
	"time"

	"github.com/promiseofcake/artifactsmmo-go-client/client"

	"github.com/promiseofcake/artifactsmmo-engine/internal/models"
)

// Response is a generic return value for most actions
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
type SkillResponse struct {
	Response
	SkillInfo client.SkillInfoSchema
}

type BankResponse struct {
	Response
	BankItems []client.SimpleItemSchema
	Item      client.ItemSchema
}

type LocationMap map[string]Location
type Locations []Location
type Location struct {
	Name string `json:"name"`
	Skin string `json:"skin"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Code string `json:"code"`
	Type string `json:"type"`
}

func locationPK(loc Location) string {
	return loc.Type + "|" + loc.Code
}

func LocationsToMap(locs Locations) LocationMap {
	locationMap := make(LocationMap)
	for _, l := range locs {
		locationMap[locationPK(l)] = l
	}
	return locationMap
}

type MonsterMap map[string]*Monster
type Monsters []Monster
type Monster struct {
	Name     string   `json:"name"`
	Skin     string   `json:"skin"`
	Code     string   `json:"code"`
	Level    int      `json:"level"`
	Location Location `json:"location"`
}

func monsterPK(monster Monster) string {
	return "monster|" + monster.Code
}

func MonstersToMap(monsters Monsters) MonsterMap {
	monsterMap := make(MonsterMap)
	for _, m := range monsters {
		monsterMap[monsterPK(m)] = &m
	}
	return monsterMap
}

func (m MonsterMap) FindMonsters(l LocationMap) {
	for _, v := range m {
		if loc, ok := l[monsterPK(*v)]; ok {
			v.Location = loc
		}
	}
}
