package main

import (
	"cart/game"
)

//go:export update
func update() {
	game.UpdateTitle()
	game.Update()
	game.Render()
}
