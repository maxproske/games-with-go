package main

import (
	"fmt"

	. "github.com/maxproske/games-with-go/15_abstract_syntax_trees/apt"
)

func main() {
	x := &OpX{}
	y := &OpY{}
	plus := &OpPlus{}
	sin := &OpSin{}
	sin.Child = x
	plus.LeftChild = sin
	plus.RightChild = y

	fmt.Println(plus.Eval(5, 2))
	fmt.Println(plus) // Println will know to use our String function

}
