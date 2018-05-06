package apt

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/PrawnSkunk/games-with-go/10_package_noise"
)

//////////////////////
// Node
//////////////////////

// An interface is a contract that lets you define a set of functions
type Node interface {
	// Provide an eval function. When given a node of a tree (+), it will compute the full result
	Eval(x, y float32) float32
	String() string
	AddRandom(node Node)                   // add to a random location in the tree
	NodeCounts() (nodeCount, nilCount int) // get back # nodes, and # nil pointers (+/sin ops without leaf nodes)
}

//////////////////////
// LeafNode
//////////////////////

// Expressions with 0 children (eg. x, y, constants)
type LeafNode struct{}

func (leaf *LeafNode) AddRandom(node Node) {
	// Cannot add anything to a leaf node
	//panic("ERROR: You tried to add a node to a leaf node")
	fmt.Println("Warning: tried to add a node to a leaf node.")
}

func (leaf *LeafNode) NodeCounts() (nodeCount, nilCount int) {
	return 1, 0 // no nil pointers
}

//////////////////////
// SingleNode
//////////////////////

// Expressions with 1 child (eg. sin, cos)
type SingleNode struct {
	Child Node
}

func (single *SingleNode) AddRandom(node Node) {
	if single.Child == nil {
		single.Child = node // no child has been assigned
	} else {
		// child has already been assigned, pass the node along using recursion
		single.Child.AddRandom(node)
	}
}

func (single *SingleNode) NodeCounts() (nodeCount, nilCount int) {
	if single.Child == nil {
		return 1, 1 // return ourself and a child
	} else {
		childNodeCount, childNilCount := single.Child.NodeCounts()
		return 1 + childNodeCount, childNilCount // ourselves+count
	}
}

//////////////////////
// DoubleNode
//////////////////////

// Expressions with 2 children (eg. +, -)
type DoubleNode struct {
	LeftChild  Node
	RightChild Node
}

func (double *DoubleNode) AddRandom(node Node) {
	// We have a choice of adding it to the right or left side
	r := rand.Intn(2) // 0-1
	if r == 0 {
		// Do left side
		if double.LeftChild == nil {
			double.LeftChild = node
		} else {
			double.LeftChild.AddRandom(node)
		}
	} else {
		// Else do right side
		if double.RightChild == nil {
			double.RightChild = node
		} else {
			double.RightChild.AddRandom(node)
		}
	}
}

func (double *DoubleNode) NodeCounts() (nodeCount, nilCount int) {
	var leftCount, leftNilCount, rightCount, rightNilCount int
	if double.LeftChild == nil {
		// No left child
		leftNilCount = 1
		leftCount = 0
	} else {
		// There is a left child
		leftCount, leftNilCount = double.LeftChild.NodeCounts()
	}

	if double.RightChild == nil {
		// No right child
		rightNilCount = 1
		rightCount = 0
	} else {
		// There is a right child
		rightCount, rightNilCount = double.RightChild.NodeCounts()
	}
	return 1 + leftCount + rightCount, leftNilCount + rightNilCount // +1 is ourselves
}

//////////////////////
// OpX
//////////////////////

// Not equivalent to "type OpX LeafNode", because it doesn't inherit LeafNode
type OpX struct {
	LeafNode
}

func (op *OpX) Eval(x, y float32) float32 {
	return x // base case when stopping recursion
}

func (op *OpX) String() string {
	return "X"
}

//////////////////////
// OpY
//////////////////////

type OpY struct {
	LeafNode
}

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

//////////////////////
// Functions
//////////////////////

func GetRandomNode() Node {
	// Non-leaf nodes only
	r := rand.Intn(9)
	switch r {
	case 0:
		return &OpPlus{}
	case 1:
		return &OpMinus{}
	case 2:
		return &OpMult{}
	case 3:
		return &OpDiv{}
	case 4:
		return &OpAtan2{}
	case 5:
		return &OpAtan{}
	case 6:
		return &OpCos{}
	case 7:
		return &OpSin{}
	case 8:
		return &OpNoise{}
	}
	panic("Get Random Node Failed")
}

func GetRandomLeaf() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return &OpX{}
	case 1:
		return &OpY{}
	case 2:
		return &OpConstant{LeafNode{}, rand.Float32()*2 - 1} // Between 0-1
	}
	panic("Get Random Leaf Failed")
}
