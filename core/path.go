package core

import (
	"fmt"
	"math"
)

// Simply a coordinate pair to show the placement of a Cell
type Coordinates struct {
	X, Y int
}

func (c Coordinates) String() string {
	return string(fmt.Sprintf("{%v, %v}", c.X, c.Y))
}

// Linked list to represent a route
type Route struct {
	Coords Coordinates
	Next   *Route
	Prev   *Route
}

// Formatted error for usage in Route
func (Route) Error(s string) error {
	return fmt.Errorf("Route error: %v", s)
}

// Check that coordinates are within bounds
func (c Coordinates) IsValid(maxX, maxY uint) bool {
	return c.X >= 0 && c.X <= int(maxX) && c.Y >= 0 && c.Y <= int(maxY)
}

// Calculate distance between two coordinates
func (c Coordinates) Distance(dest Coordinates) float64 {
	return math.Sqrt(math.Pow(float64(dest.X)-float64(c.X), 2) + math.Pow(float64(dest.Y)-float64(c.Y), 2))
}

// Return a copy of the chosen route until its n-th step
func (r Route) CopyUntil(n uint) (Route, error) {
	new_route := Route{Coords: r.Coords}
	new_route_head := &new_route
	for i := uint(0); i < n; i++ {
		if r.Next != nil {
			new_route.Next = &Route{Coords: r.Next.Coords}
			new_route, r = *new_route.Next, *r.Next
		} else if i+1 < n {
			return Route{}, r.Error(fmt.Sprintf("n=%v is bigger than route size\n", n))
		}
	}

	return *new_route_head, nil
}

// Convert route to a string representation
func (r Route) String() string {
	repr_string := r.Coords.String()
	for r.Next != nil {
		r = *r.Next
		repr_string += fmt.Sprintf(", %v", r)
	}

	return repr_string
}
