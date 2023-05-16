package core

import (
	"math"
)

// Simply a coordinate pair to show the placement of a Cell
type Coordinates struct {
	X, Y int
}

type Route struct {
	Coords Coordinates
	Next   *Route
	Prev   *Route
}

func (c Coordinates) IsValid(maxX, maxY uint) bool {
	return c.X >= 0 && c.X <= int(maxX) && c.Y >= 0 && c.Y <= int(maxY)
}

func (c Coordinates) Distance(dest Coordinates) float64 {
	return math.Sqrt(math.Pow(float64(dest.X)-float64(c.X), 2) + math.Pow(float64(dest.Y)-float64(c.Y), 2))
}