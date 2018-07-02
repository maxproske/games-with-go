package main

// https://youtu.be/Jy919y3ezOI?t=1346

import (
	"github.com/maxproske/games-with-go/24_movement/game"
	"github.com/maxproske/games-with-go/24_movement/ui2d"
)

func main() {
	ui := &ui2d.UI2d{}
	game.Run(ui)
}
