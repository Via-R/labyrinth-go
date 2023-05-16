// Main labyrinth functionality
package core

// Value to represent what is placed in the labyrinth
type Cell uint

// Simply a coordinate pair to show the placement of a Cell
type Coordinates struct {
	X, Y uint
}

// Enum for possible things placed in a cell
const (
	Empty Cell = iota
	Wall
	Start
	Finish
	Path
)

// String representation of a singular Cell
func (c *Cell) String() string {
	switch *c {
	case Empty:
		return "∘"
	case Wall:
		return "■"
	case Start:
		return "s"
	case Finish:
		return "f"
	case Path:
		return "x"
	default:
		return "?"
	}
}

// Container for Labyrinth and additional characteristics
type Field struct {
	Labyrinth     [][]Cell
	Width, Length uint
}

// Change the size of labyrinth
// Clears up all cells
func (f *Field) SetSize(width, length uint) {
	f.Labyrinth = make([][]Cell, length)
	for i := range f.Labyrinth {
		f.Labyrinth[i] = make([]Cell, width)
		for j := range f.Labyrinth[i] {
			f.Labyrinth[i][j] = Empty
		}
	}
	f.Width, f.Length = width, length
}

// Clear up all cells
func (f *Field) MakeEmpty() {
	for i := range f.Labyrinth {
		for j := range f.Labyrinth[i] {
			f.Labyrinth[i][j] = Empty
		}
	}
}

// Set start and finish points
func (f *Field) SetStartAndFinish(start, finish Coordinates) {
	f.Labyrinth[start.Y][start.X] = Start
	f.Labyrinth[finish.Y][finish.X] = Finish
}

// Create a string representation of a labyrinth row
func cellsArrayToString(cells []Cell, delimeter string) string {
	row_string := ""
	for i := range cells {
		row_string += " " + cells[i].String()
	}
	return row_string
}

// String representation of the entire labyrinth and its data
func (f Field) String() string {
	field_string := "\n"

	for i := len(f.Labyrinth) - 1; i >= 0; i-- {
		field_string += cellsArrayToString(f.Labyrinth[i], " ") + "\n"
	}

	return field_string
}
