package main

import (
	"fmt"
	builder "github.com/Via-R/labyrinth-go/builder"
	core "github.com/Via-R/labyrinth-go/core"
)

const build_new_labyrinth = false

// Show a freshly generated labyrinth
func builderDemo(l *core.Field) {
	l.SetSize(16, 16)
	l.SetStartAndFinish(core.Coordinates{X: 0, Y: 4}, core.Coordinates{X: 15, Y: 3})
	if err := builder.GenerateLabyrinth(l); err != nil {
		panic(err)
	}
	fmt.Println(l)
	if err := l.SaveLabyrinthToFile("examples/16x16.json"); err != nil {
		fmt.Println(err)
	}
}

// Load an existing labyrinth and solve it
func solverDemo(l *core.Field) {
	if err := l.LoadLabyrinthFromFile("examples/16x16.json"); err != nil {
		panic(err)
	}
	fmt.Println(l)
}

func main() {
	fmt.Println("Labyrinth sandbox")
	var l core.Field
	if err := l.Init("config.toml"); err != nil {
		panic(err)
	}

	if build_new_labyrinth {
		builderDemo(&l)
	} else {
		solverDemo(&l)
	}
}
