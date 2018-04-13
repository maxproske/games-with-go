package main

/**
 * TODO:
 * 1. Add NPCs you can talk to
 * 2. NPCs move around
 * 3. Items that can be picked up or placed down (eg. start becomes lit after picking up lantern)
 * 4. Accept natural language as input (accept key verbs)
 */

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type storyNode struct {
	text    string
	choices *choices // linked list, but there is a better data structure to use for performance
}

type choices struct {
	cmd         string
	description string
	nextNode    *storyNode
	nextChoice  *choices
}

var scanner *bufio.Scanner // make a scanner once

func main() {
	scanner = bufio.NewScanner(os.Stdin)

	// Build a simple story
	start := storyNode{text: `You are in a large chamber, deep underground.
		You see three passages leading out. A north passage leads into darkness.
		To the south, a passage appears to head upward. The eastern passages appears
		flat and well travelled.`} // new syntax, out of order, and without providing all!

	darkRoom := storyNode{text: "It is pitch black. You cannot see a thing."}
	darkRoomLit := storyNode{text: "The dark passage is now lit by your lantern. You can continue north or head back south."}
	grue := storyNode{text: "While stumbling around in the darkness, you are eaten by a grue."} // Zork reference, woo!
	trap := storyNode{text: "You head down the well travelled path when suddenly, a trap door opens and you fall into a pit."}
	treasure := storyNode{text: "You arrive at a small chamber, filled with treasures. You win!"} // win condition

	start.addChoice("N", "Go North", &darkRoom)
	start.addChoice("S", "Go South", &darkRoom) // Goes to the same place! Magic!
	start.addChoice("E", "Go East", &trap)

	darkRoom.addChoice("S", "Try to go back south", &grue)
	darkRoom.addChoice("O", "Turn on lantern", &darkRoomLit)

	darkRoomLit.addChoice("N", "Continue North", &treasure)
	darkRoomLit.addChoice("S", "Go back South", &start)

	// Start the text adventure
	start.play()

	// Run when game is completed
	fmt.Println()
	fmt.Println("The End.")
}

// Add a choice to the graph
// Pass command, description, and the next node
func (node *storyNode) addChoice(cmd string, description string, nextNode *storyNode) {

	// Make this choice. Next choice will start out nil
	choice := &choices{cmd, description, nextNode, nil} // short hand to get address right away

	if node.choices == nil {
		// Set our new choice right away
		node.choices = choice
	} else {
		// Loop down choice linked list, until we find the last item
		currentChoice := node.choices
		for currentChoice.nextChoice != nil {
			currentChoice = currentChoice.nextChoice
		}
		// We are at the last item in the list. Add to end.
		currentChoice.nextChoice = choice
	}
}

// Print out description of a room, with all available choices
func (node *storyNode) render() {
	fmt.Println(node.text)
	currentChoice := node.choices

	for currentChoice != nil {
		// Print the command
		fmt.Println(currentChoice.cmd, ":", currentChoice.description)
		// Advance to the next choice
		currentChoice = currentChoice.nextChoice
	}
}

// See if user typed command is valid. Operate on a story node.
// Return story node we will continue to.
func (node *storyNode) executeCmd(cmd string) *storyNode {
	currentChoice := node.choices
	for currentChoice != nil {
		// Make commands case insensitive
		if strings.ToLower(currentChoice.cmd) == strings.ToLower(cmd) {
			// Send user to next node
			return currentChoice.nextNode
		}
		// See if command matches any of our choices
		currentChoice = currentChoice.nextChoice
	}
	fmt.Println("Sorry, I didn't understand that.")
	return node // user won't go anywhere, return user to same node
}

func (node *storyNode) play() {

	// In order to play, we need to render the current node
	node.render()
	// If there are any options
	if node.choices != nil {
		// Read from standard input
		scanner.Scan()
		node.executeCmd(scanner.Text()).play() // recursively play
	}
}
