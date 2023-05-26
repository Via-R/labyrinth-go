package main

import (
	"fmt"
	core "github.com/Via-R/labyrinth-go/core"
)

func main() {
	fmt.Println("Labyrinth sandbox")
	l := core.Field{}
	l.SetSize(16, 16)
	l.SetStartAndFinish(core.Coordinates{X: 0, Y: 4}, core.Coordinates{X: 15, Y: 3})
	err := l.GenerateLabyrinth(1, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(l)
}
