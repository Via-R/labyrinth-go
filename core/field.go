// Main labyrinth functionality
package core

import (
	"fmt"
)

// Container for labyrinth and additional characteristics
type Field struct {
	labyrinth     [][]cell
	Width, Length uint
	Start, Finish Coordinates
	Configuration *configuration
}

// Return serialized labyrinth data
func (f *Field) GetLabyrinth() [][]uint {
	serialized_data := make([][]uint, f.Length)
	for row_idx, row := range f.labyrinth {
		serialized_data[row_idx] = make([]uint, f.Width)
		for cell_idx, cell := range row {
			serialized_data[row_idx][cell_idx] = uint(cell)
		}
	}

	return serialized_data
}

// Load labyrinth from input data
func (f *Field) LoadLabyrinth(l [][]uint) error {
	if len(l) == 0 {
		return f.Error("Cannot load empty array as a labyrinth")
	}
	width, length := len(l[0]), len(l)
	labyrinth := make([][]cell, len(l))
	var start, finish *Coordinates
	for row_idx, row := range l {
		if len(row) != width {
			return f.Error(fmt.Sprintf("Array should be rectangular, first row had %v elements, and row #%v has %v", width, row_idx, len(row)))
		}
		labyrinth[row_idx] = make([]cell, width)
		for cell_idx, cell_data := range row {
			if cell(cell_data) >= Unknown {
				return f.Error(fmt.Sprintf("Cannot use %v as a cell value", cell_data))
			}
			new_cell := cell(cell_data)
			labyrinth[row_idx][cell_idx] = new_cell
			switch new_cell {
			case Start:
				start = &Coordinates{X: cell_idx, Y: row_idx}
			case Finish:
				finish = &Coordinates{X: cell_idx, Y: row_idx}
			}
		}
	}
	if start == nil || finish == nil {
		return f.Error("No start and/or finish in the data")
	}

	f.Width, f.Length, f.labyrinth, f.Start, f.Finish = uint(width), uint(length), labyrinth, *start, *finish

	return nil
}

// Set up configuration values from .toml file
func (f *Field) Init(filename string) error {
	var config configuration
	if err := config.LoadFromFile(filename); err != nil {
		return err
	}
	f.Configuration = &config

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
		f.labyrinth[f.Start.Y][f.Start.X] = Start
		f.labyrinth[f.Finish.Y][f.Finish.X] = Finish
	} else {
		f.Start, f.Finish = Coordinates{-1, -1}, Coordinates{-1, -1}
	}
}

// Get cell type at given coordinates
func (f *Field) At(c Coordinates) (cell, error) {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return Empty, f.Error(fmt.Sprintf("Cannot get cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}

	return f.labyrinth[c.Y][c.X], nil
}

// Change cell type at the chosen coordinates
func (f *Field) Set(new_cell cell, c Coordinates) error {
	if !c.IsValid(f.Width-1, f.Length-1) {
		return f.Error(fmt.Sprintf("Cannot set cell %v out of field's bounds w=%v l=%v", c, f.Width, f.Length))
	}
	if old_cell, _ := f.At(c); old_cell != Start && old_cell != Finish {
		f.labyrinth[c.Y][c.X] = new_cell
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
	field_string := fmt.Sprintf("Size: %vx%v\nStart: %v\nFinish: %v\n\n", f.Width, f.Length, f.Start, f.Finish)

	for i := len(f.labyrinth) - 1; i >= 0; i-- {
		field_string += cellsArrayToString(f.labyrinth[i], " ") + "\n"
	}

	return field_string[:len(field_string)-1]
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
func (f *Field) FillEmptyCellsWithWalls() {
	for i := range f.labyrinth {
		for j := range f.labyrinth[i] {
			coords := Coordinates{X: j, Y: i}
			if cell, err := f.At(coords); err != nil {
				continue
			} else if !cell.IsBlocking(true) {
				f.Set(Wall, coords)
			} else if cell == Path {
				f.Set(Empty, coords)
			}
		}
	}
}
