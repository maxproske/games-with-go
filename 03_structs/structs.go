package main

import (
	"fmt"
)

type position struct {
	x float32
	y float32
}

type badGuy struct {
	name   string
	health int
	pos    position
}

func main() {
	//var p position
	//pos.x = 5
	//pos.y = 4

	// New struct
	p := position{4, 2}
	fmt.Println(p.x)

	// Pass a struct to a new struct
	b := badGuy{"Jabba The Hut", 100, p}
	fmt.Println(b)

	// Pass structure to function
	whereIsBadGuy(b)
}

func whereIsBadGuy(b badGuy) {
	x := b.pos.x
	y := b.pos.y
	fmt.Println("(", x, ",", y, ")")
}
