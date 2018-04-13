package main

import (
	"bufio"
	"fmt"
	"os"
)

// Data structure for binary tree instead of linked list
type storyNode struct {
	text    string
	yesPath *storyNode
	noPath  *storyNode
}

func main() {
	root := storyNode{"You are at the entrance to a dark cave. Do you want to go in the cave?", nil, nil}
	losing := storyNode{"You entered the cave and were eaten!", nil, nil}
	winning := storyNode{"You did not enter, and avoided being eaten! You have won!", nil, nil}

	root.yesPath = &losing
	root.noPath = &winning

	root.printStory(0)
	fmt.Println("---------")
	root.play()
}

// Receiver type
func (node *storyNode) play() {
	fmt.Println(node.text)

	if node.yesPath != nil && node.noPath != nil {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			answer := scanner.Text()
			// Respond to input
			if answer == "yes" {
				node.yesPath.play()
				break
			} else if answer == "no" {
				node.noPath.play()
				break
			} else {
				fmt.Println("Please answer yes or no.")
			}
		}
	}
}

// print whole tree with recursion
func (node *storyNode) printStory(depth int) {
	// indentation of two characters
	for i := 0; i < depth; i++ {
		fmt.Print("-> ")
	}
	fmt.Println(node.text)
	if node.yesPath != nil {
		node.yesPath.printStory(depth + 1)
	}
	if node.noPath != nil {
		node.noPath.printStory(depth + 1)
	}
}
