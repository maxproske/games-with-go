package main

import "fmt"

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
	// Pass bad guy by reference
	p := position{4, 2}
	b := badGuy{"Jabba The Hut", 100, p}
	whereIsBadGuy(&b)

	// Pass as a pointer?
	// Save memory and allow its vales to be changed? Then Yes.
}

// Size of integer is size of pointer, but badGuy struct is much more!
// Pass by reference/indirectly, instead of making copies of badGuy
func whereIsBadGuy(bPtr *badGuy) {
	// Go automatically dereferences structs, C/C++ do not. They need -> to access members inside a pointer.
	// (There is nothing else you could possibly mean)
	// In Java, you always pass indirectly, it's just hidden from you
	x := bPtr.pos.x
	y := bPtr.pos.y
	fmt.Println("(", x, ",", y, ")")
}
