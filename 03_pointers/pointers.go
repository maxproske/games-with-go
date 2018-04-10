package main

import "fmt"

func main() {
	x := 5
	fmt.Println(x)

	// *int		pointer to an int
	// &x		address of x
	// *x		dereference x (value of x)

	// Get the address of x (in virtual memory)
	// var xPtr *int = &x
	xPtr := &x
	fmt.Println(xPtr)

	// Add one, and have it actually change
	addOneByReference(xPtr)
	fmt.Println(x)
}

func addOneByReference(xPtr *int) {
	// Dereference the pointer
	*xPtr = *xPtr + 1
}
