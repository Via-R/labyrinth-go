package core

import (
	"fmt"
	"math/rand"
)

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

// Find all possible choices from goven coordinates
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
	// fmt.Printf("Reverse distances: %v\n", reverse_distances)
	for i := range choices {
		distances[i] = 1 / choices[i].Distance(f.Finish)
		if reverse_distances {
			distances[i] = 1 / distances[i]
		}
		sum += distances[i]
	}
	// fmt.Printf("Distances: %v\n", distances)

	for i := range distances {
		probabilities[i] = distances[i] / sum
	}
	// fmt.Printf("Probabilities: %v\n", probabilities)
	sum = 0
	for i := range probabilities {
		probability_limits[i] = sum + probabilities[i]
		sum += probabilities[i]
	}
	// fmt.Printf("Probability limits: %v\n", probability_limits)
	choice_cursor := rand.Float64()
	// fmt.Printf("Choice cursor: %v\n", choice_cursor)
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

// Surround selected cell with walls, except for the neighboring cells to except_coords
func (f *Field) surroundWithWalls(coords Coordinates, except_coords Coordinates) {
	for _, shift := range MooreShifts {
		choice := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		// fmt.Printf("Trying to set a wall at%v: ", choice)
		if cell, err := f.at(choice); err != nil || cell != Empty || except_coords.Distance(choice) <= 1 {
			// fmt.Printf("Fail - err=%v, cell=%v, dist=%v\n", err, cell, except_coords.Distance(choice))
			continue
		} else {
			// fmt.Print("Success\n")
			f.set(Wall, choice)
		}
	}
}

func (f *Field) reachFinishOrLoop(route_tail Route) (Route, uint, bool, error) {
	route_head := &route_tail
	route_length := uint(1)

	safety_counter := 0
	const safety_limit = 100

	for route_head.Coords != f.Finish && safety_counter < safety_limit {
		choices, err := f.findChoices(route_head.Coords)
		if err != nil {
			return Route{}, 0, true, err
		}

		if len(choices) == 0 {
			return route_tail, route_length, false, f.Error("Could not reach the finish")
		}

		next_coords, err := f.selectChoice(choices, 0)
		if err = f.set(Path, next_coords); err != nil {
			return Route{}, 0, true, err
		}

		new_path_part := Route{Coords: next_coords, Prev: route_head}
		route_head.Next = &new_path_part
		route_head = &new_path_part
		route_length++

		safety_counter++
	}

	if safety_counter == safety_limit {
		return Route{}, 0, true, f.Error("Safety limit exceeded in route builder")
	}

	return route_tail, route_length, false, nil
}

type RouteWithLength struct {
	r Route
	l uint
}

// Generate labyrinth based on complexity (in percents)
func (f *Field) GenerateLabyrinth(complexity float64) error {
	if complexity < 0 || complexity > 1 {
		return f.Error("Complexity (percentage) cannot be less than 0 or over 1")
	} else if !f.Start.IsValid(f.Width-1, f.Length-1) || !f.Finish.IsValid(f.Width-1, f.Length-1) {
		return f.Error("Start and/or finish are out of bounds or not set yet")
	}

	route_start := Route{Coords: f.Start}
	safety_counter := 0
	const safety_limit = 20

	routes := make([]RouteWithLength, 0, safety_limit)

	for float64(f.CountCells()[Empty])/float64(f.Size()) > 0.5 && safety_counter < safety_limit {
		route, route_length, failure, err := f.reachFinishOrLoop(route_start)
		if failure {
			return err
		}
		routes = append(routes, RouteWithLength{route, route_length})
		if err != nil {
			fmt.Println(err)
		}
		base_route_info := routes[rand.Intn(len(routes))]
		base_route_split_idx := rand.Intn(int(base_route_info.l)) //pick random route here from routes
		new_route := Route{Coords: base_route_info.r.Coords}
		fmt.Printf("Route n=%v len=%v\n", base_route_split_idx, route_length)
		// Somewhere after this out of bounds access or nil dereference happened, maybe even in a new iteration
		ip := base_route_info.r
		for i := 0; i < base_route_split_idx; i++ {
			if ip.Next == nil {
				return f.Error(fmt.Sprintf("Failed to reach random point in route n=%v len=%v", base_route_split_idx, route_length))
			}
			ip = *ip.Next
			new_route.Next = &Route{Coords: ip.Next.Coords}
			new_route = *new_route.Next
		}

		route_start = new_route
		fmt.Println(f)
		fmt.Printf("New start: %v\n", route_start.Coords)
		safety_counter++
	}

	if safety_counter == safety_limit {
		return f.Error("Safety limit exceeded in routes generator")
	} else {
		fmt.Printf("Area filled! \ncells=%v \nempty area=%v\n", f.CountCells(), float32(f.CountCells()[Empty])/float32(f.Size())*100)
	}

	return nil
}
