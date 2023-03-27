package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)
import (
	"github.com/niftysoft/2d-platformer/internal"
	"log"
)

func main() {
	game, err := internal.NewGame()
	if err != nil {
		log.Fatal(err)
	}

	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Axolotl Knight")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
