package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type storyNode struct {
	text   string
	choice []*choice // Use slices to make text adventure more efficient, and simpler too!
}

type choice struct {
	cmd         string
	description string
	nextNode    *storyNode
}

var scanner *bufio.Scanner // make a scanner once

// Size is part of the type.
// For this reason, arrays aren't used very often in go.
// Instead, we use slices.
func printArray(a [3]string) {
	for _, e := range a {
		fmt.Print(e)
	}
}

// Notice no [3]. Slices know about their own size.
func printSlice(a []string) {
	for _, e := range a {
		fmt.Print(e)
	}
}

func main() {

	// new array that is 3 items long
	abc := [3]string{"a", "b", "c"}
	printArray(abc)

	// New slice
	abc2 := []string{"a", "b", "c"}
	// Built-in functions to make larger/smaller
	abc2 = append(abc2, "d") // makes a new array, old is garbage collected
	printSlice(abc2)

	scanner = bufio.NewScanner(os.Stdin)

	// Build a simple story
	start := storyNode{text: `You are in a large chamber, deep underground.
	You see three passages leading out. A north passage leads into darkness.
	To the south, a passage appears to head upward. The eastern passages appears
	flat and well travelled.`}

	darkRoom := storyNode{text: "It is pitch black. You cannot see a thing."}
	darkRoomLit := storyNode{text: "The dark passage is now lit by your lantern. You can continue north or head back south."}
	grue := storyNode{text: "While stumbling around in the darkness, you are eaten by a grue."}
	trap := storyNode{text: "You head down the well travelled path when suddenly, a trap door opens and you fall into a pit."}
	treasure := storyNode{text: "You arrive at a small chamber, filled with treasures. You win!"}

	// Adding is slower with slices. But playing is faster, and simpler logic.
	start.addChoice("N", "Go North", &darkRoom)
	start.addChoice("S", "Go South", &darkRoom)
	start.addChoice("E", "Go East", &trap)

	darkRoom.addChoice("S", "Try to go back south", &grue)
	darkRoom.addChoice("O", "Turn on lantern", &darkRoomLit)

	darkRoomLit.addChoice("N", "Continue North", &treasure)
	darkRoomLit.addChoice("S", "Go back South", &start)

	start.play()

	fmt.Println()
	fmt.Println("The End.")
}

// Add a choice to the graph
// Pass command, description, and the next node
func (node *storyNode) addChoice(cmd string, description string, nextNode *storyNode) {
	choice := &choice{cmd, description, nextNode} // we don't need the nil pointer anymore
	node.choice = append(node.choice, choice)
}

// Print out description of a room, with all available choice
func (node *storyNode) render() {
	fmt.Println(node.text)
	// Check if there are any choices
	if node.choice != nil {
		// New feature -- ranges
		// We don't care about the index, we just want the element (for each?)
		for _, choice := range node.choice {
			// For each choice, print
			fmt.Println(choice.cmd, choice.description)
		}
	}
}

// See if user typed command is valid.
func (node *storyNode) executeCmd(cmd string) *storyNode {
	for _, choice := range node.choice {
		if strings.ToLower(choice.cmd) == strings.ToLower(cmd) {
			return choice.nextNode
		}
	}
	fmt.Println("Sorry, I didn't understand that.")
	return node
}

func (node *storyNode) play() {

	node.render()
	if node.choice != nil {
		scanner.Scan()
		node.executeCmd(scanner.Text()).play()
	}
}
