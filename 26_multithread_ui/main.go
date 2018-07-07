package main

// https://youtu.be/Jy919y3ezOI?t=1346

import (
	"runtime"

	"github.com/maxproske/games-with-go/26_multithread_ui/game"
	"github.com/maxproske/games-with-go/26_multithread_ui/ui2d"
)

func main() {
	// Make new game
	game := game.NewGame(3, "game/maps/level1.map")

	// Make our UIs
	for i := 0; i < 3; i++ {
		go func(i int) {
			runtime.LockOSThread() // Goroutines must stay on the same thread for the window to draw and handle input
			ui := ui2d.NewUI(game.InputChan, game.LevelChans[i])
			ui.Run()
		}(i) // Loop will finish quickly, so pass i in
	}

	game.Run()
}
