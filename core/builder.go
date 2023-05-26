package core

import (
	"fmt"
	"math/rand"
)

// Arrays of all possible coordinates' shifts when going around Von Neumann's and Moore's neighborhoods
var NeumannShifts = [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
var MooreShifts = [8][2]int{{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}

// Count blocking cells around the chosen coordinates
func (f *Field) countWallsAround(coords Coordinates, finish_reached bool) uint {
	counter := uint(0)
	for _, shift := range MooreShifts {
		neighbor := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		if cell, err := f.at(neighbor); err == nil && cell.IsBlocking(finish_reached) {
			counter++
		}
	}

	return counter
}

// Check that the cell can be a part of the route with one of the available ChoiceChecker's
func (f *Field) isChoiceValid(coords Coordinates, finish_reached bool) bool {
	checker := isChoiceValidBy2CloseBlocksGetter()

	return checker(f, coords, finish_reached)
}

// Find all possible choices from given coordinates
// NOTE: it might make sense to use the "no more than N blocking cells around" in case
// we don't want to have crossing routes
func (f *Field) findChoices(coords Coordinates, finish_reached bool) ([]Coordinates, error) {
	choices := make([]Coordinates, 0, 4)
	for _, shift := range NeumannShifts {
		choice := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		if cell, err := f.at(choice); err == nil && !cell.IsBlocking(finish_reached) && f.isChoiceValid(choice, finish_reached) {
			choices = append(choices, choice)
		}
	}

	return choices, nil
}

// Select one of the choices based on distance to finish
// Choices are made based on probability, which is proportionate to the distance to finish
// Probabilities are flipped if complexity is high enough
func (f *Field) selectChoice(choices []Coordinates, complexity float64) (Coordinates, error) {
	if complexity < 0 || complexity > 1 {
		return Coordinates{-1, -1}, f.Error("Complexity (percentage) cannot be less than 0 or over 1")
	}
	switch len(choices) {
	case 0:
		return Coordinates{-1, -1}, nil
	case 1:
		return choices[0], nil
	}

	for _, choice := range choices {
		if choice == f.Finish {
			return choice, nil
		}
	}

	distances, probabilities, probability_limits := make([]float64, len(choices)), make([]float64, len(choices)), make([]float64, len(choices))
	sum := 0.
	reverse_distances := rand.Float64() < complexity
	for i := range choices {
		distances[i] = 1 / choices[i].Distance(f.Finish)
		if reverse_distances {
			distances[i] = 1 / distances[i]
		}
		sum += distances[i]
	}

	for i := range distances {
		probabilities[i] = distances[i] / sum
	}

	sum = 0
	for i := range probabilities {
		probability_limits[i] = sum + probabilities[i]
		sum += probabilities[i]
	}

	choice_cursor := rand.Float64()
	choice_idx := -1
	for i, limit := range probability_limits {
		if choice_cursor < limit {
			choice_idx = i
			break
		}
	}

	return choices[choice_idx], nil
}

// Continue the given route until it gets stuck or reaches the finish
func (f *Field) reachFinishOrLoop(route Route, complexity float64, finish_reached bool) (Route, error) {
	safety_counter := 0
	const safety_limit = 10000

	for route.End.Coords != f.Finish && safety_counter < safety_limit {
		choices, err := f.findChoices(route.End.Coords, finish_reached)
		if err != nil {
			return Route{}, err
		}

		if len(choices) == 0 {
			return route, nil
		}

		next_coords, err := f.selectChoice(choices, complexity)
		if err = f.set(Path, next_coords); err != nil {
			return Route{}, err
		}

		route.Add(next_coords)
		safety_counter++
	}

	if safety_counter == safety_limit {
		return Route{}, f.Error("Safety limit exceeded in route builder")
	}

	return route, nil
}

// Go through all provided routes and update their BaseIndices field
func (f *Field) processRoutesForBaseCompatibility(routes *[]Route, finish_reached bool) {
	for route_idx, route := range *routes {
		it := route.GetIterator()
		base_indices := make([]uint, 0, route.Length)
		route_part_idx := uint(0)
		for coords, is_end := it(); !is_end; coords, is_end = it() {
			if options, err := f.findChoices(coords, finish_reached); err == nil && len(options) > 0 {
				base_indices = append(base_indices, route_part_idx)
			}
			route_part_idx++
		}
		(*routes)[route_idx].BaseIndices = base_indices
	}
}

// Remove routes that do not have possible bases for new routes
func (f *Field) removeNonBaseRoutes(routes *[]Route) {
	idx := 0
	for _, route := range *routes {
		if len(route.BaseIndices) > 0 {
			(*routes)[idx] = route
			idx++
		}
	}
	*routes = (*routes)[:idx]
}

// Generate routes for empty labyrinth with defined start and finish cells
// Generation stops if the labyrinth is filled up to (1-max_empty_area*100%)
// Will retry the generation 'max_retries' if area is not filled up yet or fail if the 'max_retries' is reached,
// if 'only_one_path_near_finish' is true, then only one path cell can be around finish
func (f *Field) GenerateRoutes(complexity, max_empty_area float64, max_retries uint, only_one_path_near_finish bool) error {
	safety_counter := uint(0)
	routes := make([]Route, 1, max_retries)
	routes[0] = Route{}
	routes[0].Init(f.Start)
	finish_reached := false

	for float64(f.CountCells()[Empty])/float64(f.Size()) > max_empty_area && safety_counter < max_retries {
		// update BaseIndices for all routes to show which route parts can be bases for new routes
		f.processRoutesForBaseCompatibility(&routes, finish_reached)
		// remove routes that cannot provide any new routes
		f.removeNonBaseRoutes(&routes)

		if len(routes) == 0 {
			fmt.Println(float64(f.CountCells()[Empty]) / float64(f.Size()))
			return f.Error("Cannot form new routes but area is not filled yet")
		}

		// pick one of the routes
		base_route := routes[rand.Intn(len(routes))]

		// create a copy until one of the possible bases and kick off a new route
		base_route_split_idx := base_route.BaseIndices[rand.Intn(len(base_route.BaseIndices))] // pick random route that has possible bases
		new_route_base, err := base_route.CopyUntil(uint(base_route_split_idx) + 1)
		if err != nil {
			return err
		}

		new_route, err := f.reachFinishOrLoop(new_route_base, complexity, finish_reached)
		if err != nil {
			return err
		}
		if end_cell, err := f.at(new_route.End.Coords); err != nil {
			return err
		} else if only_one_path_near_finish && end_cell == Finish {
			finish_reached = true
		}

		routes = append(routes, new_route)
		safety_counter++
	}
	if !finish_reached {
		return f.Error("Area filled but finish was not reached")
	}
	if safety_counter == max_retries {
		fmt.Println("Safety limit exceeded in routes generator")
	} else {
		fmt.Printf("Area filled! \ncells=%v \nempty area=%v\n", f.CountCells(), float32(f.CountCells()[Empty])/float32(f.Size())*100)
	}

	return nil
}

// Generate labyrinth based on complexity (in percents)
func (f *Field) GenerateLabyrinth(complexity float64, only_one_path_near_finish bool) error {
	if !f.Start.IsValid(f.Width-1, f.Length-1) || !f.Finish.IsValid(f.Width-1, f.Length-1) {
		return f.Error("Start and/or finish are out of bounds or not set yet")
	}

	const safety_limit, min_area = 10, 0.5
	route_retries := f.Size()

	err := f.GenerateRoutes(complexity, min_area, route_retries, only_one_path_near_finish)
	safety_counter := 0
	for ; err != nil && safety_counter < safety_limit; safety_counter++ {
		f.MakeEmpty(true)
		err = f.GenerateRoutes(complexity, min_area, route_retries, only_one_path_near_finish)
	}

	fmt.Println(err)
	if safety_counter == safety_limit {
		return f.Error("Safety limit exceeded in labyrinth generator")
	}

	f.fillEmptyCellsWithWalls()

	return nil
}
