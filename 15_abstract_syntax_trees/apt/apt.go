package apt

import (
	"math"
)

// An interface is a contract that lets you define a set of functions
type Node interface {
	// Provide an eval function. When given a node of a tree (+), it will compute the full result
	Eval(x, y float32) float32
	String() string
}

// Expressions with 0 children (eg. x, y, constants)
type LeafNode struct{}

// Expressions with 1 child (eg. sin, cos)
type SingleNode struct {
	Child Node
}

// Expressions with 2 children (eg. +, -)
type DoubleNode struct {
	LeftChild  Node
	RightChild Node
}

//////////////////////
// OpX
//////////////////////

// Equivalent to: type OpX struct { LeafNode }
type OpX LeafNode

func (op *OpX) Eval(x, y float32) float32 {
	return x // base case when stopping recursion
}

func (op *OpX) String() string {
	return "X"
}

//////////////////////
// OpY
//////////////////////

type OpY LeafNode

func (op *OpY) Eval(x, y float32) float32 {
	return y
}

func (op *OpY) String() string {
	return "Y"
}

//////////////////////
// OpPlus
//////////////////////

type OpPlus struct {
	// Struct embedding. Instead of giving it a LeftChild and RightChild node, say it has a DoubleNode
	DoubleNode
}

// OpPlus.Eval(...) has the same signature as the Node interface
func (op *OpPlus) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) + op.RightChild.Eval(x, y)
}

func (op *OpPlus) String() string {
	return "( + " + op.LeftChild.String() + " " + op.RightChild.String() + " )"
}

//////////////////////
// OpSin
//////////////////////

type OpSin struct {
	SingleNode
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Child.Eval(x, y))))
}

func (op *OpSin) String() string {
	return "( Sin " + op.Child.String() + " )" // use recursion to print child tree
}
