package apt

import (
	"math"
	"strconv"

	"github.com/maxproske/games-with-go/10_package_noise"
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
// OpConstant
//////////////////////

type OpConstant struct {
	LeafNode
	value float32
}

func (op *OpConstant) Eval(x, y float32) float32 {
	return op.value
}

func (op *OpConstant) String() string {
	return strconv.FormatFloat(float64(op.value), 'f', 9, 32) // Converts floating point to a string
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
// OpMinus
//////////////////////

type OpMinus struct {
	DoubleNode
}

func (op *OpMinus) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) - op.RightChild.Eval(x, y)
}

func (op *OpMinus) String() string {
	return "( - " + op.LeftChild.String() + " " + op.RightChild.String() + " )"
}

//////////////////////
// OpMult
//////////////////////

type OpMult struct {
	DoubleNode
}

func (op *OpMult) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) * op.RightChild.Eval(x, y)
}

func (op *OpMult) String() string {
	return "( * " + op.LeftChild.String() + " " + op.RightChild.String() + " )"
}

//////////////////////
// OpDiv
//////////////////////

type OpDiv struct {
	DoubleNode
}

func (op *OpDiv) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) / op.RightChild.Eval(x, y)
}

func (op *OpDiv) String() string {
	return "( / " + op.LeftChild.String() + " " + op.RightChild.String() + " )"
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
	return "( Sin " + op.Child.String() + " )"
}

//////////////////////
// OpCos
//////////////////////

type OpCos struct {
	SingleNode
}

func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Child.Eval(x, y))))
}

func (op *OpCos) String() string {
	return "( Cos " + op.Child.String() + " )"
}

//////////////////////
// OpAtan
//////////////////////

type OpAtan struct {
	SingleNode
}

func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Child.Eval(x, y))))
}

func (op *OpAtan) String() string {
	return "( Atan " + op.Child.String() + " )"
}

//////////////////////
// OpAtan2
//////////////////////

type OpAtan2 struct {
	DoubleNode
}

func (op *OpAtan2) Eval(x, y float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}

func (op *OpAtan2) String() string {
	return "( Atan2 " + op.LeftChild.String() + " " + op.RightChild.String() + " )"
}

//////////////////////
// OpNoise
//////////////////////

type OpNoise struct {
	DoubleNode
}

func (op *OpNoise) Eval(x, y float32) float32 {
	// (80*) makes it between 0-2
	// (- 2.0) maks it between -1 and 1
	return 80*noise.Snoise2(op.LeftChild.Eval(x, y), op.RightChild.Eval(x, y)) - 2.0
}

func (op *OpNoise) String() string {
	return "( SimplexNoise " + op.LeftChild.String() + " " + op.RightChild.String() + " )"
}
