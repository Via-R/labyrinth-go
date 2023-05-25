package core

import (
	"fmt"
)

// Linked list to represent a route
type Route struct {
	Start, End  *routeStep
	Length      uint
	BaseIndices []uint
}

// Linked list node
type routeStep struct {
	Coords Coordinates
	Next   *routeStep
	Prev   *routeStep
}

// String representation of one route step
func (r routeStep) String() string {
	return fmt.Sprintf("Step %v, next=%p, prev=%p", r.Coords, r.Next, r.Prev)
}

// Formatted error for usage in Route
func (Route) Error(s string) error {
	return fmt.Errorf("Route error: %v", s)
}

// Initialize a route with one step
func (r *Route) Init(c Coordinates) error {
	if r.Start != nil || r.End != nil || r.Length != 0 {
		return r.Error("Route.Init can only be called once, some of the fields were already initialized")
	}

	new_node := routeStep{Coords: c, Next: nil, Prev: nil}
	r.Start, r.End, r.Length = &new_node, &new_node, 1

	return nil
}

// Add a new routeStep to the route
func (r *Route) Add(c Coordinates) error {
	if r.Start == nil || r.End == nil || r.Length == 0 {
		return r.Error("Route.Init needs to be called first, none of the fields were initialized")
	}
	new_node := routeStep{Coords: c, Next: nil, Prev: r.End}
	r.End.Next = &new_node
	r.End = &new_node
	r.Length++

	return nil
}

// Return iterator to go over the entire Route
func (r *Route) GetIterator() func() (Coordinates, bool) {
	curr_node := r.Start
	return func() (Coordinates, bool) {
		if curr_node == nil {
			return Coordinates{}, true
		}
		defer func() { curr_node = curr_node.Next }()
		return curr_node.Coords, false
	}
}

// String representation of Route
func (r Route) String() string {
	if r.Length == 0 {
		return "[]"
	}
	repr_string := "[ "
	it := r.GetIterator()
	const separator = " -> "
	for coords, is_end := it(); !is_end; coords, is_end = it() {
		repr_string += coords.String() + separator
	}

	return repr_string[:len(repr_string)-len(separator)] + " ]"
}

// Return a copy of the chosen route with n steps
func (r Route) CopyUntil(n uint) (Route, error) {
	if r.Length == 0 {
		return Route{}, r.Error("Cannot copy an uninitialized route")
	} else if n > r.Length {
		return Route{}, r.Error(fmt.Sprintf("n=%v is bigger than route length=%v\n", n, r.Length))
	}

	it := r.GetIterator()
	new_route := Route{}
	coords, is_end := it()
	new_route.Init(coords)
	if is_end {
		return new_route, nil
	}
	for counter := uint(1); counter < n; counter++ {
		coords, _ = it()
		new_route.Add(coords)
	}

	return new_route, nil
}
