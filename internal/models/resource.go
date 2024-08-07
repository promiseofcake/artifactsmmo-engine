package models

import "github.com/promiseofcake/artifactsmmo-go-client/client"

type ResourceMap map[string]*Resource
type Resources []Resource

// Resource is the struct for interacting with resources on the map
// For now there is only ever one kind of resource for a given code
type Resource struct {
	Name     string                     `json:"name"`
	Code     string                     `json:"code"`
	Skill    client.ResourceSchemaSkill `json:"skill"`
	Level    int                        `json:"level"`
	Location Location                   `json:"location"`
}

// GetCoords returns a Monster's coords
func (r Resource) GetCoords() Coords {
	return Coords{
		X: r.Location.Coords.X,
		Y: r.Location.Coords.Y,
	}
}

func resourcePK(resource Resource) string {
	return "resource|" + resource.Code
}

func ResourcesToMap(resources Resources) ResourceMap {
	monsterMap := make(ResourceMap)
	for _, m := range resources {
		monsterMap[resourcePK(m)] = &m
	}
	return monsterMap
}

func (r ResourceMap) FindResources(l LocationMap) {
	for _, v := range r {
		if loc, ok := l[resourcePK(*v)]; ok {
			v.Location = loc
		}
	}
}

func (r ResourceMap) ToSlice() Resources {
	var resources Resources
	for _, m := range r {
		resources = append(resources, *m)
	}
	return resources
}
