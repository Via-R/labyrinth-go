package core

// Value to represent what is placed in the labyrinth
type cell uint

// Enum for possible things placed in a cell
const (
	Empty cell = iota
	Wall
	Start
	Finish
	Path
	Unknown // should always be last for type validation
)

// String representation of a singular Cell
func (c cell) String() string {
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

// Check if the cell cannot be a part of the route
func (c cell) IsBlocking(finish_is_blocking bool) bool {
	return c == Wall || c == Path || c == Start || c == Finish && finish_is_blocking
}

// Create a string representation of a labyrinth row
func cellsArrayToString(cells []cell, delimeter string) string {
	row_string := ""
	for _, cell := range cells {
		row_string += " " + cell.String()
	}
	return row_string
}
