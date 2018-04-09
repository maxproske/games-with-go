package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Make a scanner
	s := bufio.NewScanner(os.Stdin)

	// Get user input
	fmt.Println("What is your name?")
	s.Scan()
	name := s.Text()

	// Greet the user
	sayHelloTo(name)
	s.Scan()
	fmt.Println("Bye.")
}

func sayHelloTo(name string) {
	fmt.Println("Hello,", name+".")
}
