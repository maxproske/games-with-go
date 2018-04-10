/**
 * Games With Go EP 03.2 - Functions, Structs, Pointers
 * https://www.youtube.com/watch?v=039Ma0MzUE4
 */

package main

import "fmt"

// Linked list. Series of nodes (text, pointer to next node)
type storyPage struct {
	text     string
	nextPage *storyPage // avoid recursion
}

func main() {
	page1 := storyPage{"It was a dark and stormy night.", nil} // nil is vaguely similar to null in other languages
	page2 := storyPage{"You are alone.", nil}
	page3 := storyPage{"You see a troll ahead.", nil} // will let us know when we've reached the end of the story
	page1.nextPage = &page2
	page2.nextPage = &page3

	readPage(&page1)
}

func readPage(page *storyPage) {
	if page == nil {
		return
	}
	fmt.Println("(", &page, ")", page.text)
	readPage(page.nextPage)
}
