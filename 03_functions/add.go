package main

import "fmt"

func main() {
	// Use return values
	x := 5
	x = addOne(x)
	fmt.Println(x)

	// Compose functions
	x = addOne(addOne(x))
	fmt.Println(x)
}

func addOne(x int) int {
	return x + 1
}
