package builder

import (
	"fmt"
	core "github.com/Via-R/labyrinth-go/core"
	"math/rand"
)

// Arrays of all possible coordinates' shifts when going around Von Neumann's and Moore's neighborhoods
var NeumannShifts = [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
var MooreShifts = [8][2]int{{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}

// Count blocking cells around the chosen coordinates
func countWallsAround(f *core.Field, coords core.Coordinates, finish_reached bool) uint {
	counter := uint(0)
	for _, shift := range MooreShifts {
		neighbor := core.Coordinates{X: coords.X + shift[0], Y: coords.Y + shift[1]}
		if cell, err := f.At(neighbor); err == nil && cell.IsBlocking(finish_reached) {
			counter++
		}
	}

	return counter
}

// Check that the cell can be a part of the route with one of the available ChoiceChecker's
func isChoiceValid(f *core.Field, coords core.Coordinates, finish_reached bool) bool {
	checker := isChoiceValidBy2CloseBlocksGetter()

	return checker(f, coords, finish_reached)
}

// Find all possible choices from given coordinates
// NOTE: it might make sense to use the "no more than N blocking cells around" in case
// we don't want to have crossing routes
func findChoices(f *core.Field, coords core.Coordinates, finish_reached bool) ([]core.Coordinates, error) {
	choices := make([]core.Coordinates, 0, 4)
	for _, shift := range NeumannShifts {
		choice := core.Coordinates{X: coords.X + shift[0], Y: coords.Y + shift[1]}
		if cell, err := f.At(choice); err == nil && !cell.IsBlocking(finish_reached) && isChoiceValid(f, choice, finish_reached) {
			choices = append(choices, choice)
		}
	}

	return choices, nil
}

// Select one of the choices based on distance to finish
// Choices are made based on probability, which is proportionate to the distance to finish
// Probabilities are flipped if complexity is high enough
func selectChoice(f *core.Field, choices []core.Coordinates) (core.Coordinates, error) {
	switch len(choices) {
	case 0:
		return core.Coordinates{X: -1, Y: -1}, f.Error("Cannot make a choice out of zero length array")
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
	reverse_distances := rand.Float64()*100 < f.Configuration.Builder.Complexity
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
func reachFinishOrLoop(f *core.Field, route core.Route, finish_reached bool) (core.Route, error) {
	safety_counter := 0
	const safety_limit = 10000

	for route.End.Coords != f.Finish && safety_counter < safety_limit {
		choices, err := findChoices(f, route.End.Coords, finish_reached)
		if err != nil {
			return core.Route{}, err
		}

		if len(choices) == 0 {
			return route, nil
		}

		next_coords, err := selectChoice(f, choices)
		if err = f.Set(core.Path, next_coords); err != nil {
			return core.Route{}, err
		}

		route.Add(next_coords)
		safety_counter++
	}

	if safety_counter == safety_limit {
		return core.Route{}, f.Error("Safety limit exceeded in route builder")
	}

	return route, nil
}

// Go through all provided routes and update their BaseIndices field
func processRoutesForBaseCompatibility(f *core.Field, routes *[]core.Route, finish_reached bool) {
	for route_idx, route := range *routes {
		it := route.GetIterator()
		base_indices := make([]uint, 0, route.Length)
		route_part_idx := uint(0)
		for coords, is_end := it(); !is_end; coords, is_end = it() {
			if options, err := findChoices(f, coords, finish_reached); err == nil && len(options) > 0 {
				base_indices = append(base_indices, route_part_idx)
			}
			route_part_idx++
		}
		(*routes)[route_idx].BaseIndices = base_indices
	}
}

// Remove routes that do not have possible bases for new routes
func removeNonBaseRoutes(routes *[]core.Route) {
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
func generateRoutes(f *core.Field) error {
	safety_counter := uint(0)
	max_route_builds := f.Size()
	routes := make([]core.Route, 1, max_route_builds)
	routes[0] = core.Route{}
	routes[0].Init(f.Start)
	finish_reached := false

	for float64(f.CountCells()[core.Empty])/float64(f.Size())*100 > f.Configuration.Builder.MaxAreaToCoverWithWalls && safety_counter < max_route_builds {
		// update BaseIndices for all routes to show which route parts can be bases for new routes
		processRoutesForBaseCompatibility(f, &routes, finish_reached)
		// remove routes that cannot provide any new routes
		removeNonBaseRoutes(&routes)

		if len(routes) == 0 {
			return f.Error(fmt.Sprintf("Cannot form new routes but area is not filled yet (empty area=%v%%)", float64(f.CountCells()[core.Empty])/float64(f.Size())*100))
		}

		// pick one of the routes
		base_route := routes[rand.Intn(len(routes))]

		// create a copy until one of the possible bases and kick off a new route
		base_route_split_idx := base_route.BaseIndices[rand.Intn(len(base_route.BaseIndices))] // pick random route that has possible bases
		new_route_base, err := base_route.CopyUntil(uint(base_route_split_idx) + 1)
		if err != nil {
			return err
		}

		new_route, err := reachFinishOrLoop(f, new_route_base, finish_reached)
		if err != nil {
			return err
		}
		if end_cell, err := f.At(new_route.End.Coords); err != nil {
			return err
		} else if f.Configuration.Builder.OnlyOnePathNearFinish && end_cell == core.Finish {
			finish_reached = true
		}

		routes = append(routes, new_route)
		safety_counter++
	}
	if !finish_reached {
		return f.Error("Area filled but finish was not reached")
	}
	if safety_counter == max_route_builds {
		fmt.Println("Safety limit exceeded in routes generator")
	} else {
		fmt.Printf("Area filled! \ncells=%v \nempty area=%v\n", f.CountCells(), float32(f.CountCells()[core.Empty])/float32(f.Size())*100)
	}

	return nil
}

// Generate labyrinth based on configuration parameters
func GenerateLabyrinth(f *core.Field) error {
	if f.Configuration == nil {
		return f.Error("Configuration was not initialized yet")
	}
	if !f.Start.IsValid(f.Width-1, f.Length-1) || !f.Finish.IsValid(f.Width-1, f.Length-1) {
		return f.Error("Start and/or finish are out of bounds or not set yet")
	}

	err := generateRoutes(f)
	safety_counter := uint(0)
	for ; err != nil && safety_counter < f.Configuration.Builder.LabyrinthBuilderAtempts; safety_counter++ {
		f.MakeEmpty(true)
		err = generateRoutes(f)
	}

	if safety_counter == f.Configuration.Builder.LabyrinthBuilderAtempts {
		return f.Error("Safety limit exceeded in labyrinth generator")
	}

	f.FillEmptyCellsWithWalls()

	return nil
}
