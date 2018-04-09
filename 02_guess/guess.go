package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Make a scanner
	scanner := bufio.NewScanner(os.Stdin)

	low := 1
	high := 100

	fmt.Println("Please think of a number between", low, "and", high)
	fmt.Println("Press ENTER when ready.")

	// Get input from user
	scanner.Scan()

loop:
	for i, previousGuess := 0, 0; ; i++ {
		guess := (low + high) / 2

		// User is lying
		if guess == previousGuess || guess == 0 {
			fmt.Println("You're cheating!")
			break loop
		}

		fmt.Println("I guess the number is", guess)
		fmt.Println("Is that:")
		fmt.Println("(a) too high?")
		fmt.Println("(b) too low?")
		fmt.Println("(c) correct?")

		// Get input from user
		scanner.Scan()
		response := scanner.Text()

		// Binary search
		switch response {
		case "a":
			high = guess - 1
		case "b":
			low = guess + 1
		case "c":
			fmt.Println("I won in", (i + 1), "tries!")
			break loop
		default:
			fmt.Print("Invalid response, try again.")
		}

		// Set last response
		previousGuess = guess
	}
}
