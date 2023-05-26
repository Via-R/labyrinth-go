package main

import (
	"fmt"
	builder "github.com/Via-R/labyrinth-go/builder"
	core "github.com/Via-R/labyrinth-go/core"
)

func main() {
	fmt.Println("Labyrinth sandbox")
	var l core.Field
	if err := l.Init("config.toml"); err != nil {
		panic(err)
	}
	l.SetSize(16, 16)
	l.SetStartAndFinish(core.Coordinates{X: 0, Y: 4}, core.Coordinates{X: 15, Y: 3})
	if err := builder.GenerateLabyrinth(&l); err != nil {
		panic(err)
	}
	fmt.Println(l)
}
