package core

import (
	"fmt"
	"math/rand"
)

// Arrays of all possible coordinates' shifts when going around Von Neumann's and Moore's neighborhoods
var NeumannShifts = [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
var MooreShifts = [8][2]int{{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}

// Count blocking cells around the chosen coordinates
func (f *Field) countWallsAround(coords Coordinates) uint {
	counter := uint(0)
	for _, shift := range MooreShifts {
		neighbor := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		if cell, err := f.at(neighbor); err == nil && cell.IsBlocking() {
			counter++
		}
	}

	return counter
}

// Check that the Moore's neighborhood of the cell at the chosen coordinates doesn't have any corners made of 3 blocking cells
func (f *Field) isChoiceValid(coords Coordinates) bool {
	thirds_counter := 0
	corner_dots_counter := 0
	loopedMooreShifts := append(MooreShifts[:], MooreShifts[0])
	for i := 0; i < len(loopedMooreShifts); i++ {
		thirds_counter++
		shift := loopedMooreShifts[i]
		choice := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		if cell, err := f.at(choice); err == nil && cell.IsBlocking() {
			corner_dots_counter++
		}
		if corner_dots_counter == 3 {
			return false
		}
		if thirds_counter == 3 {
			thirds_counter = 0
			corner_dots_counter = 0
			i--
		}
	}

	return true
}

// Find all possible choices from given coordinates
func (f *Field) findChoices(coords Coordinates) ([]Coordinates, error) {
	choices := make([]Coordinates, 0, 4)
	for _, shift := range NeumannShifts {
		choice := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		if cell, err := f.at(choice); err == nil && !cell.IsBlocking() && f.isChoiceValid(choice) {
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
	// fmt.Printf("Choice idx: %v\n", choice_idx)

	return choices[choice_idx], nil
}

// Continue the given route until it gets stuck or reaches the finish
func (f *Field) reachFinishOrLoop(route Route, complexity float64) (Route, error) {
	safety_counter := 0
	const safety_limit = 100

	for route.End.Coords != f.Finish && safety_counter < safety_limit {
		choices, err := f.findChoices(route.End.Coords)
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

// Generate routes for empty labyrinth with defined start and finish cells
// Generation stops if the labyrinth is filled up to (1-max_empty_area*100%)
// Will retry the generation 'max_retries' if area is not filled up yet or fail if the 'max_retries' is reached
func (f *Field) GenerateRoutes(complexity, max_empty_area float64, max_retries uint) (*[]Route, error) {
	safety_counter := uint(0)
	routes := make([]Route, 0, max_retries)
	route := Route{}
	route.Init(f.Start)

	for float64(f.CountCells()[Empty])/float64(f.Size()) > 0.4 && safety_counter < max_retries {
		new_route, err := f.reachFinishOrLoop(route, complexity)
		if err != nil {
			return nil, err
		}

		routes = append(routes, new_route) // TODO: think of a way not to add the same routes twice here
		for _, r := range routes {
			fmt.Println(r)
		}

		base_route := routes[rand.Intn(len(routes))]
		base_route_split_idx := rand.Intn(int(base_route.Length)) // pick random route here from 'routes'

		route, err = base_route.CopyUntil(uint(base_route_split_idx))
		if err != nil {
			return nil, err
		}

		safety_counter++
	}

	if safety_counter == max_retries {
		fmt.Println("Safety limit exceeded in routes generator")
		return nil, nil
	} else {
		fmt.Printf("Area filled! \ncells=%v \nempty area=%v\n", f.CountCells(), float32(f.CountCells()[Empty])/float32(f.Size())*100)
	}

	return &routes, nil
}

// Generate labyrinth based on complexity (in percents)
func (f *Field) GenerateLabyrinth(complexity float64) error {
	if !f.Start.IsValid(f.Width-1, f.Length-1) || !f.Finish.IsValid(f.Width-1, f.Length-1) {
		return f.Error("Start and/or finish are out of bounds or not set yet")
	}

	const safety_limit, min_area, route_retries = 3, 0.4, 10
	safety_counter := 0

	routes, err := f.GenerateRoutes(complexity, min_area, route_retries)
	for ; routes == nil && safety_counter < safety_limit; safety_counter++ {
		if err != nil {
			return err
		}
		f.MakeEmpty(true)
		routes, err = f.GenerateRoutes(complexity, min_area, route_retries)
	}

	if safety_counter == safety_limit {
		return f.Error("Safety limit exceeded in labyrinth generator")
	}

	fmt.Println("Winning routes:")
	for _, r := range *routes {
		fmt.Println(r)
	}

	// TODO: if the labyrinth is generated normally, fill the empty area with walls and remove Paths

	return nil
}
