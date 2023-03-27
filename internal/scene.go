package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	Update() error
	Draw(*ebiten.Image)
	Layout(int, int) (int, int)
}

// BaseScene takes over rendering the root scene of a Game.
type BaseScene struct {
	game *Game
}

// NewBaseScene constructs a new scene which has a reference to the provided Game.
func NewBaseScene(g *Game) *BaseScene {
	return &BaseScene{game: g}
}

// Update updates this scene every tick.
func (s *BaseScene) Update() error {
	return nil
}

// Draw draws the game screen; called every frame.
func (s *BaseScene) Draw(screen *ebiten.Image) {

}

// Layout maps the window size to the screen size.
func (s *BaseScene) Layout(w, h int) (int, int) {
	return 320, 240
}
