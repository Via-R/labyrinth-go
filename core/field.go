// Main labyrinth functionality
package core

import (
	"fmt"
	"math/rand"
)

// Container for labyrinth and additional characteristics
type Field struct {
	labyrinth     [][]Cell
	Width, Length uint
	Start, Finish Coordinates
	Solution      Route
}

// Change the size of labyrinth
// Clears up all cells
func (f *Field) SetSize(width, length uint) {
	f.labyrinth = make([][]Cell, length)
	for i := range f.labyrinth {
		f.labyrinth[i] = make([]Cell, width)
	}
	f.Width, f.Length = width, length
	f.MakeEmpty()
}

// Clear up all cells
func (f *Field) MakeEmpty() {
	for i := range f.labyrinth {
		for j := range f.labyrinth[i] {
			f.labyrinth[i][j] = Empty
		}
	}
	f.Start, f.Finish = Coordinates{-1, -1}, Coordinates{-1, -1}
}

// Get cell type at given coordinates
func (f *Field) at(c Coordinates) (Cell, error) {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return Empty, f.Error(fmt.Sprintf("Cannot get cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}

	return f.labyrinth[c.Y][c.X], nil
}

// Change cell type at the chosen coordinates
func (f *Field) set(cell Cell, c Coordinates) error {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return f.Error(fmt.Sprintf("Cannot set cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}
	if old_cell, _ := f.at(c); old_cell != Start && old_cell != Finish {
		f.labyrinth[c.Y][c.X] = cell
	}

	return nil
}

// Set start and finish points
func (f *Field) SetStartAndFinish(start, finish Coordinates) {
	f.labyrinth[start.Y][start.X] = Start
	f.labyrinth[finish.Y][finish.X] = Finish
	f.Start, f.Finish = start, finish
}

// String representation of the entire labyrinth and its data
func (f Field) String() string {
	field_string := "\n"

	for i := len(f.labyrinth) - 1; i >= 0; i-- {
		field_string += cellsArrayToString(f.labyrinth[i], " ") + "\n"
	}

	return field_string
}

// Formatted error for usage in Field
func (Field) Error(s string) error {
	return fmt.Errorf("labyrinth error: %v", s)
}

// Find all possible choices from goven coordinates
func (f *Field) findChoices(coords Coordinates) ([]Coordinates, error) {
	shifts := [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
	choices := make([]Coordinates, 0, 4)
	for _, shift := range shifts {
		choice := Coordinates{coords.X + shift[0], coords.Y + shift[1]}
		if cell, err := f.at(choice); err != nil {
			continue
		} else if !cell.IsBlocking() {
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
	shifts := [8][2]int{{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}

	for _, shift := range shifts {
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

// Generate solution based on complexity (in percents)
func (f *Field) GenerateSolution(complexity float64) error {
	if complexity < 0 || complexity > 1 {
		return f.Error("Complexity (percentage) cannot be less than 0 or over 1")
	} else if !f.Start.IsValid(f.Width-1, f.Length-1) || !f.Finish.IsValid(f.Width-1, f.Length-1) {
		return f.Error("Start and/or finish are out of bounds or not set yet")
	}

	route_tail := Route{Coords: f.Start}
	route_head := &route_tail

	counter := 0

	for route_head.Coords != f.Finish && counter < 40 {
		choices, err := f.findChoices(route_head.Coords)
		if err != nil {
			return err
		}

		if len(choices) == 0 {
			return f.Error("Could not reach the finish")
		}

		next_coords, err := f.selectChoice(choices, complexity)
		if err = f.set(Path, next_coords); err != nil {
			return err
		}

		f.surroundWithWalls(route_head.Coords, next_coords)

		new_path_part := Route{Coords: next_coords, Prev: route_head}
		route_head.Next = &new_path_part
		route_head = &new_path_part

		counter++
	}
	f.surroundWithWalls(f.Finish, Coordinates{-1, -1})

	return nil
}
