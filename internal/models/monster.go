package models

type MonsterMap map[string]*Monster
type Monsters []Monster

type Monster struct {
	Name     string   `json:"name"`
	Skin     string   `json:"skin"`
	Code     string   `json:"code"`
	Level    int      `json:"level"`
	Location Location `json:"location"`
}

// GetCoords returns a Monster's coords
func (m Monster) GetCoords() Coords {
	return Coords{
		X: m.Location.Coords.X,
		Y: m.Location.Coords.Y,
	}
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
