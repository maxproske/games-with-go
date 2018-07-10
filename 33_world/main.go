package main

// https://youtu.be/Jy919y3ezOI?t=1346

import (
	"github.com/maxproske/games-with-go/33_world/game"
	"github.com/maxproske/games-with-go/33_world/ui2d"
)

func main() {
	// Make new game
	game := game.NewGame(1)
	go func() {
		game.Run()
	}()

	// Make our UI
	ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	ui.Run()
}
