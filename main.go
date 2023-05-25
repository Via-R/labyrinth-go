package main

import (
	"fmt"
	core "github.com/Via-R/labyrinth-go/core"
)

func main() {
	fmt.Println("Labyrinth sandbox")
	l := core.Field{}
	l.SetSize(42, 42)
	l.SetStartAndFinish(core.Coordinates{X: 0, Y: 4}, core.Coordinates{X: 41, Y: 2})
	err := l.GenerateLabyrinth(0, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(l)
}
