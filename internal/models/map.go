package models

import "math"

// Coords defines X, Y map coordinates as signed integers
type Coords struct {
	X int
	Y int
}

// CalculateDistance determines the number of moves (distance) to get
// from one Coords to a second Coords using the Manhattan distance forumla
func CalculateDistance(one, two Coords) int {
	return int(math.Abs(float64(one.X-two.X)) + math.Abs(float64(one.Y-two.Y)))
}
