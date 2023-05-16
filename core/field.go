// Main labyrinth functionality
package core

import (
	"fmt"
	"math/rand"
)

// Container for Labyrinth and additional characteristics
type Field struct {
	Labyrinth     [][]Cell
	Width, Length uint
	Start, Finish Coordinates
	Solution      Route
}

// Change the size of labyrinth
// Clears up all cells
func (f *Field) SetSize(width, length uint) {
	f.Labyrinth = make([][]Cell, length)
	for i := range f.Labyrinth {
		f.Labyrinth[i] = make([]Cell, width)
	}
	f.Width, f.Length = width, length
	f.MakeEmpty()
}

// Clear up all cells
func (f *Field) MakeEmpty() {
	for i := range f.Labyrinth {
		for j := range f.Labyrinth[i] {
			f.Labyrinth[i][j] = Empty
		}
	}
	f.Start, f.Finish = Coordinates{-1, -1}, Coordinates{-1, -1}
}

// Get cell type at given coordinates
func (f *Field) at(c Coordinates) (Cell, error) {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return Empty, f.Error(fmt.Sprintf("Cannot get cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}

	return f.Labyrinth[c.Y][c.X], nil
}

// Set start and finish points
func (f *Field) SetStartAndFinish(start, finish Coordinates) {
	f.Labyrinth[start.Y][start.X] = Start
	f.Labyrinth[finish.Y][finish.X] = Finish
	f.Start, f.Finish = start, finish
}

// String representation of the entire labyrinth and its data
func (f Field) String() string {
	field_string := "\n"

	for i := len(f.Labyrinth) - 1; i >= 0; i-- {
		field_string += cellsArrayToString(f.Labyrinth[i], " ") + "\n"
	}

	return field_string
}

// Formatted error for usage in Field
func (Field) Error(s string) error {
	return fmt.Errorf("Labyrinth error: %v", s)
}

// Find all possible choices from goven coordinates
func (f *Field) findChoices(coords Coordinates) ([]Coordinates, error) {
	shifts := [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
	choices := make([]Coordinates, 0, 4)
	for i := range shifts {
		choice := Coordinates{coords.X + shifts[i][0], coords.Y + shifts[i][1]}
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

	distances, probabilities, probability_limits := make([]float64, len(choices)), make([]float64, len(choices)), make([]float64, len(choices))
	sum := 0.
	reverse_distances := rand.Float64() < complexity
	fmt.Printf("Reverse distances: %v\n", reverse_distances)
	for i := range choices {
		distances[i] = 1 / choices[i].Distance(f.Finish)
		if reverse_distances {
			distances[i] = 1 / distances[i]
		}
		sum += distances[i]
	}
	fmt.Printf("Distances: %v\n", distances)

	for i := range distances {
		probabilities[i] = distances[i] / sum
	}
	fmt.Printf("Probabilities: %v\n", probabilities)
	sum = 0
	for i := range probabilities {
		probability_limits[i] = sum + probabilities[i]
		sum += probabilities[i]
	}
	fmt.Printf("Probability limits: %v\n", probability_limits)
	choice_cursor := rand.Float64()
	fmt.Printf("Choice cursor: %v\n", choice_cursor)
	choice_idx := -1
	for i := range probability_limits {
		if choice_cursor < probability_limits[i] {
			choice_idx = i
			break
		}
	}
	fmt.Printf("Choice idx: %v\n", choice_idx)

	return choices[choice_idx], nil
}

// Generate solution based on complexity (in percents)
func (f *Field) GenerateSolution(complexity float64) error {
	if complexity < 0 || complexity > 1 {
		return f.Error("Complexity (percentage) cannot be less than 0 or over 1")
	} else if !f.Start.IsValid(f.Width-1, f.Length-1) || !f.Finish.IsValid(f.Width-1, f.Length-1) {
		return f.Error("Start and/or finish are out of bounds or not set yet")
	}

	route := Route{Coords: f.Start}
	route_head := &route
	choices, err := f.findChoices(route.Coords)
	counter := 0

	for len(choices) > 0 && route_head.Coords != f.Finish && counter < 1 {
		if err != nil {
			return err
		}

		fmt.Println(choices)
		f.selectChoice(choices, complexity)
		counter++
	}

	return nil
}
