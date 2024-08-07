package models

type LocationMap map[string]Location
type Locations []Location
type Location struct {
	Name   string `json:"name"`
	Skin   string `json:"skin"`
	Coords Coords `json:"coords"`
	Code   string `json:"code"`
	Type   string `json:"type"`
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
