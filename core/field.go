// Main labyrinth functionality
package core

import (
	"fmt"
)

// Container for labyrinth and additional characteristics
type Field struct {
	labyrinth     [][]cell
	Width, Length uint
	start, finish Coordinates
	configuration *configuration
}

// Set up configuration values from .toml file
func (f *Field) Init(filename string) error {
	var config configuration
	if err := config.LoadFromFile(filename); err != nil {
		return err
	}
	f.configuration = &config

	return nil
}

// Change the size of labyrinth
// Clears up all cells
func (f *Field) SetSize(width, length uint) {
	f.labyrinth = make([][]cell, length)
	for i := range f.labyrinth {
		f.labyrinth[i] = make([]cell, width)
	}
	f.Width, f.Length = width, length
	f.MakeEmpty(false)
}

// Clear up all cells except for start and finish if the flag is true
func (f *Field) MakeEmpty(leave_start_and_finish bool) {
	for i := range f.labyrinth {
		for j := range f.labyrinth[i] {
			f.labyrinth[i][j] = Empty
		}
	}
	if leave_start_and_finish {
		f.labyrinth[f.start.Y][f.start.X] = Start
		f.labyrinth[f.finish.Y][f.finish.X] = Finish
	} else {
		f.start, f.finish = Coordinates{-1, -1}, Coordinates{-1, -1}
	}
}

// Get cell type at given coordinates
func (f *Field) at(c Coordinates) (cell, error) {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return Empty, f.Error(fmt.Sprintf("Cannot get cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}

	return f.labyrinth[c.Y][c.X], nil
}

// Change cell type at the chosen coordinates
func (f *Field) set(new_cell cell, c Coordinates) error {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return f.Error(fmt.Sprintf("Cannot set cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}
	if old_cell, _ := f.at(c); old_cell != Start && old_cell != Finish {
		f.labyrinth[c.Y][c.X] = new_cell
	}

	return nil
}

// Set start and finish points
func (f *Field) SetStartAndFinish(start, finish Coordinates) {
	f.labyrinth[start.Y][start.X] = Start
	f.labyrinth[finish.Y][finish.X] = Finish
	f.start, f.finish = start, finish
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

// Count all cell types in the labyrinth
func (f *Field) CountCells() map[cell]uint {
	counter := make(map[cell]uint)

	for _, row := range f.labyrinth {
		for _, cell := range row {
			counter[cell]++
		}
	}

	return counter
}

// Count the amount of all cells in labyrinth
func (f *Field) Size() uint {
	return f.Width * f.Length
}

// Replace paths with empty cells and fill the rest with walls
func (f *Field) fillEmptyCellsWithWalls() {
	for i := range f.labyrinth {
		for j := range f.labyrinth[i] {
			coords := Coordinates{X: j, Y: i}
			if cell, err := f.at(coords); err != nil {
				continue
			} else if !cell.IsBlocking(true) {
				f.set(Wall, coords)
			} else if cell == Path {
				f.set(Empty, coords)
			}
		}
	}
}
