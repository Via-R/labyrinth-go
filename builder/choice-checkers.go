package builder

import (
	core "github.com/Via-R/labyrinth-go/core"
)

type ChoiceChecker func(f *core.Field, coords core.Coordinates, finish_reached bool) bool

// Check that the Moore's neighborhood of the cell at the chosen coordinates doesn't have any corners made of 3 blocking cells
func isChoiceValidByCornersGetter() ChoiceChecker {
	return func(f *core.Field, coords core.Coordinates, finish_reached bool) bool {
		thirds_counter := 0
		corner_dots_counter := 0
		loopedMooreShifts := append(MooreShifts[:], MooreShifts[0])
		for i := 0; i < len(loopedMooreShifts); i++ {
			thirds_counter++
			shift := loopedMooreShifts[i]
			choice := core.Coordinates{X: coords.X + shift[0], Y: coords.Y + shift[1]}
			cell, err := f.At(choice)

			if err == nil && finish_reached && cell == core.Finish {
				// if we want only one path near finish, we eliminate choices that are in moore's neighborhood with 'Finish' cell
				return false
			}
			if err == nil && cell.IsBlocking(finish_reached) {
				corner_dots_counter++
			}
			if corner_dots_counter == 3 {
				return false
			}
			if thirds_counter == 3 {
				thirds_counter = 0
				corner_dots_counter = 0
				i--
			}
		}

		return true
	}
}

// Check that the Moore's neighborhood of the cell at the chosen coordinates doesn't have more than 'max_blocks_around' blocking cells
func isChoiceValidByNBlocksGetter(max_blocks_around uint) ChoiceChecker {
	return func(f *core.Field, coords core.Coordinates, finish_reached bool) bool {
		blocks_around := uint(0)
		for _, shift := range MooreShifts {
			choice := core.Coordinates{X: coords.X + shift[0], Y: coords.Y + shift[1]}
			cell, err := f.At(choice)

			if err == nil && finish_reached && cell == core.Finish {
				// if we want only one path near finish, we eliminate choices that are in moore's neighborhood with 'Finish' cell
				return false
			}
			if err == nil && cell.IsBlocking(finish_reached) {
				blocks_around++
			}
			if blocks_around > max_blocks_around {
				return false
			}
		}

		return true
	}
}

// Check that the route either goes straight or turns but never goes in the vicinity of other routes, including itself
func isChoiceValidBy2CloseBlocksGetter() ChoiceChecker {
	return func(f *core.Field, coords core.Coordinates, finish_reached bool) bool {
		blocks_around := 0
		var first_block *core.Coordinates = nil
		for _, shift := range MooreShifts {
			choice := core.Coordinates{X: coords.X + shift[0], Y: coords.Y + shift[1]}
			cell, err := f.At(choice)

			if err == nil && finish_reached && cell == core.Finish {
				// if we want only one path near finish, we eliminate choices that are in moore's neighborhood with 'Finish' cell
				return false
			}
			if err == nil && cell.IsBlocking(finish_reached) {
				if first_block == nil {
					first_block = &choice
				} else if first_block.Distance(choice) > 1 {
					return false
				}
				blocks_around++
			}
			if blocks_around > 2 {
				return false
			}
		}

		return true
	}
}
