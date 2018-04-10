/**
 * Games With Go EP 04 - Methods, Recursion, Linked Lists, Branching Story
 * https://www.youtube.com/watch?v=jXFZW11-M4U
 */

package main

import "fmt"

type storyPage struct {
	text     string
	nextPage *storyPage
	prevPage *storyPage // Doubley linked list (1<->2)

	// Don't put methods here!
}

func main() {
	page1 := storyPage{"It was a dark and stormy night.", nil, nil} // nil is vaguely similar to null in other languages
	page2 := storyPage{"You are alone.", nil, nil}
	page3 := storyPage{"You see a troll ahead.", nil, nil} // will let us know when we've reached the end of the story
	page1.nextPage = &page2
	page2.nextPage = &page3

	page1.readPageLoop()
}

// Reciever acts like a method (function of a class)
// No large difference behind the scenes (more of a convenience)
func (page *storyPage) readPageRecursive() {
	if page == nil {
		return
	}
	fmt.Println("(", &page, ")", page.text)

	// Compilers use tail call elimintation to automatically turn recursion into a loop
	// But Go DOES NOT do this! So we can see how deep we are in the debugger.
	// Don't depend on the compiler. Change this to use a loop so it's effective on any compiler.
	// Millions = loop
	// Short = recursion

	// No large difference behind the scenes (more of a convenience)
	page.nextPage.readPageRecursive() // Stack will grow each recursive call...storyPage
}

// Go has no tail call elimination, so loop
func (page *storyPage) readPageLoop() {
	for page != nil {
		fmt.Println("(", &page, ")", page.text)
		page = page.nextPage
	}
}

// Functions - return value
// Procedures - no return value, just executes commands
// Methods - functions attached to a struct/object
