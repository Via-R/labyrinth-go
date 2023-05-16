package main

import (
	"fmt"
	core "github.com/Via-R/labyrinth-go/core"
)

func main() {
	fmt.Println("Labyrinth sandbox")
	l := core.Field{}
	l.SetSize(8, 8)
	l.SetStartAndFinish(core.Coordinates{X: 0, Y: 4}, core.Coordinates{X: 7, Y: 2})
	fmt.Println(l)
	err := l.GenerateSolution(1)
	if err != nil {
		fmt.Println(err)
	}
}
