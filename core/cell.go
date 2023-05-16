package core

// Value to represent what is placed in the labyrinth
type Cell uint

// Enum for possible things placed in a cell
const (
	Empty Cell = iota
	Wall
	Start
	Finish
	Path
)

// String representation of a singular Cell
func (c Cell) String() string {
	switch c {
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

func (c Cell) IsBlocking() bool {
	return c == Wall || c == Path
}

// Create a string representation of a labyrinth row
func cellsArrayToString(cells []Cell, delimeter string) string {
	row_string := ""
	for _, cell := range cells {
		row_string += " " + cell.String()
	}
	return row_string
}
