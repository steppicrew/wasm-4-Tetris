package main

import (
	"cart/game"
)

//go:export update
func update() {
	game.Update()
	game.Render()
}
