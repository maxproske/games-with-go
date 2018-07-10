package main

// https://youtu.be/Jy919y3ezOI?t=1346

import (
	"github.com/maxproske/games-with-go/32_fog_of_war/game"
	"github.com/maxproske/games-with-go/32_fog_of_war/ui2d"
)

func main() {
	// Make new game
	game := game.NewGame(1, "game/maps/level1.map")
	go func() {
		game.Run()
	}()

	// Make our UI
	ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	ui.Run()
}
