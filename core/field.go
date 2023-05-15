package core

type Cell int

type Coordinates struct {
	X, Y uint
}

const (
	Empty Cell = iota
	Wall
	Start
	Finish
	Path
)

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

type Field struct {
	Labyrinth     [][]Cell
	Width, Length uint
}

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

func (f *Field) MakeEmpty() {
	for i := range f.Labyrinth {
		for j := range f.Labyrinth[i] {
			f.Labyrinth[i][j] = Empty
		}
	}
}

func (f *Field) SetStartAndFinish(start, finish Coordinates) {
	f.Labyrinth[start.Y][start.X] = Start
	f.Labyrinth[finish.Y][finish.X] = Finish
}

func cellsArrayToString(cells []Cell, delimeter string) string {
	row_string := ""
	for i := range cells {
		row_string += " " + cells[i].String()
	}
	return row_string
}

func (f Field) String() string {
	field_string := "\n"

	for i := range f.Labyrinth {
		field_string += cellsArrayToString(f.Labyrinth[i], " ") + "\n"
	}

	return field_string
}
