package apt

import (
	"math"
	"math/rand"
	"reflect"
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
	SetParent(parent Node)
	SetChildren([]Node)
	GetParent() Node
	GetChildren() []Node // We don't have the ability to access .Children
	AddRandom(node Node) // add to a random location in the tree
	AddLeaf(leaf Node) bool
	NodeCount() int
}

func Mutate(node Node) Node {
	// Get a random node to mutate into (can be regular or leaf)
	r := rand.Intn(12)
	var mutatedNode Node
	if r <= 9 {
		mutatedNode = GetRandomNode()
	} else {
		mutatedNode = GetRandomLeaf()
	}

	// Check if root of the tree
	if node.GetParent() != nil {
		// Adjust the parent's child to point to the mutated node
		for i, parentChild := range node.GetParent().GetChildren() {
			if parentChild == node {
				// We found the spot in the parent's children array to modify
				node.GetParent().GetChildren()[i] = mutatedNode
			}
		}
	}

	// Take the children for the old node, and put them in the new node
	for i, child := range node.GetChildren() {
		if i >= len(mutatedNode.GetChildren()) {
			break // If the new node doesn't have as many as the old node, break out
		}
		mutatedNode.GetChildren()[i] = child // If we do have space, set new node to spot of old node
		child.SetParent(mutatedNode)         // Set its parent
	}

	for i, child := range mutatedNode.GetChildren() {
		// It's possible the new tree hasn't got all nil pointers filled up
		if child == nil {
			leaf := GetRandomLeaf()
			leaf.SetParent(mutatedNode)
			mutatedNode.GetChildren()[i] = leaf
		}
	}

	mutatedNode.SetParent(node.GetParent())
	return mutatedNode
}

func GetRandomNode() Node {
	// Non-leaf nodes only
	r := rand.Intn(9)
	switch r {
	case 0:
		return NewOpPlus()
	case 1:
		return NewOpMinus()
	case 2:
		return NewOpMult()
	case 3:
		return NewOpDiv()
	case 4:
		return NewOpAtan2()
	case 5:
		return NewOpAtan()
	case 6:
		return NewOpCos()
	case 7:
		return NewOpSin()
	case 8:
		return NewOpNoise()
	}
	panic("Get Random Node Failed")
}

func GetRandomLeaf() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return NewOpX()
	case 1:
		return NewOpY()
	case 2:
		return NewOpConstant() // Between 0-1
	}
	panic("Get Random Leaf Failed")
}

//////////////////////
// BaseNode
//////////////////////
type BaseNode struct {
	Parent   Node
	Children []Node
}

func (node *BaseNode) AddRandom(nodeToAdd Node) {
	// Sin gets mutated to +. Now needs another branch!
	addIndex := rand.Intn(len(node.Children)) // Select which child branch we wnat to add a new node to
	if node.Children[addIndex] == nil {
		nodeToAdd.SetParent(node) // Set parent
		node.Children[addIndex] = nodeToAdd
	} else {
		node.Children[addIndex].AddRandom(nodeToAdd) // pass it down the tree
	}
}

func (node *BaseNode) SetParent(parent Node) {
	node.Parent = parent
}

func (node *BaseNode) SetChildren(children []Node) {
	node.Children = children
}

func (node *BaseNode) AddLeaf(leaf Node) bool {
	// Look for the nil children, and fill them in with leafs
	for i, child := range node.Children {
		if child == nil {
			leaf.SetParent(node)
			node.Children[i] = leaf
			return true
		} else if node.Children[i].AddLeaf(leaf) {
			return true
		}
	}
	return false // Unable to find a leaf to add to
}

// BaseNode will not be a Node until we satisfy the interface
func (node *BaseNode) Eval(x, y float32) float32 {
	panic("Tried to call eval on basenode")
}

func (node *BaseNode) String() string {
	panic("Tried to call string on basenode")
}

func (node *BaseNode) NodeCount() int {
	count := 1 // The node we are at right now
	for _, child := range node.Children {
		count += child.NodeCount()
	}
	return count
}

func (node *BaseNode) GetChildren() []Node {
	return node.Children
}

func (node *BaseNode) GetParent() Node {
	return node.Parent
}

//////////////////////
// OpX
//////////////////////

// Not equivalent to "type OpX BaseNode", because it doesn't inherit BaseNode
type OpX struct {
	BaseNode
}

// No built-in structure for constructors. But there is a convention:
func NewOpX() *OpX {
	return &OpX{BaseNode{nil, make([]Node, 0)}}
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
	BaseNode
}

func NewOpY() *OpY {
	return &OpY{BaseNode{nil, make([]Node, 0)}}
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
	BaseNode
	value float32
}

func NewOpConstant() *OpConstant {
	return &OpConstant{BaseNode{nil, make([]Node, 0)}, rand.Float32()*2 - 1}
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
	// Struct embedding. Instead of giving it a LeftChild and RightChild node, say it has a BaseNode
	BaseNode
}

func NewOpPlus() *OpPlus {
	return &OpPlus{BaseNode{nil, make([]Node, 2)}}
}

// OpPlus.Eval(...) has the same signature as the Node interface
func (op *OpPlus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) + op.Children[1].Eval(x, y)
}

func (op *OpPlus) String() string {
	return "( + " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

//////////////////////
// OpMinus
//////////////////////

type OpMinus struct {
	BaseNode
}

func NewOpMinus() *OpMinus {
	return &OpMinus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMinus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) - op.Children[1].Eval(x, y)
}

func (op *OpMinus) String() string {
	return "( - " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

//////////////////////
// OpMult
//////////////////////

type OpMult struct {
	BaseNode
}

func NewOpMult() *OpMult {
	return &OpMult{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMult) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) * op.Children[1].Eval(x, y)
}

func (op *OpMult) String() string {
	return "( * " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

//////////////////////
// OpDiv
//////////////////////

type OpDiv struct {
	BaseNode
}

func NewOpDiv() *OpDiv {
	return &OpDiv{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpDiv) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) / op.Children[1].Eval(x, y)
}

func (op *OpDiv) String() string {
	return "( / " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

//////////////////////
// OpSin
//////////////////////

type OpSin struct {
	BaseNode
}

func NewOpSin() *OpSin {
	return &OpSin{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Children[0].Eval(x, y))))
}

func (op *OpSin) String() string {
	return "( Sin " + op.Children[0].String() + " )"
}

//////////////////////
// OpCos
//////////////////////

type OpCos struct {
	BaseNode
}

func NewOpCos() *OpCos {
	return &OpCos{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Children[0].Eval(x, y))))
}

func (op *OpCos) String() string {
	return "( Cos " + op.Children[0].String() + " )"
}

//////////////////////
// OpAtan
//////////////////////

type OpAtan struct {
	BaseNode
}

func NewOpAtan() *OpAtan {
	return &OpAtan{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Children[0].Eval(x, y))))
}

func (op *OpAtan) String() string {
	return "( Atan " + op.Children[0].String() + " )"
}

//////////////////////
// OpAtan2
//////////////////////

type OpAtan2 struct {
	BaseNode
}

func NewOpAtan2() *OpAtan2 {
	return &OpAtan2{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpAtan2) Eval(x, y float32) float32 {
	return float32(math.Atan2(float64(op.Children[0].Eval(x, y)), float64(op.Children[1].Eval(x, y))))
}

func (op *OpAtan2) String() string {
	return "( Atan2 " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

//////////////////////
// OpNoise
//////////////////////

type OpNoise struct {
	BaseNode
}

func NewOpNoise() *OpNoise {
	return &OpNoise{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpNoise) Eval(x, y float32) float32 {
	// (80*) makes it between 0-2
	// (- 2.0) maks it between -1 and 1
	return 80*noise.Snoise2(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y)) - 2.0
}

func (op *OpNoise) String() string {
	return "( SimplexNoise " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

// Pass any node, make a distinct copy of the entire tree underneath
func CopyTree(node Node, parent Node) Node {
	// We can use reflection to ask the type of a struct
	copy := reflect.New(reflect.ValueOf(node).Elem().Type()).Interface().(Node) // Eg. Make a new node that is an OpPlus. Wrap back up in a node interface. Cast interface as node

	// All constants will be 0, so they won't be all black
	switch n := node.(type) {
	case *OpConstant:
		copy.(*OpConstant).value = n.value
	}

	// Blank node. No values (parents and children) not filled in.
	copy.SetParent(parent)
	copyChildren := make([]Node, len(node.GetChildren())) // Same size as children we are copying
	copy.SetChildren(copyChildren)
	for i := range copyChildren {
		copyChildren[i] = CopyTree(node.GetChildren()[i], copy) // We don't just want to copy the child. Use recursion to get children of children of children...
	}
	return copy
}

// Old node gets replaced by new node
func ReplaceNode(old Node, new Node) {
	oldParent := old.GetParent() // Take the old node's parent, and set that to a variable for convenience
	if oldParent != nil {
		// Link parent of old node to new node
		for i, child := range oldParent.GetChildren() {
			if child == old {
				oldParent.GetChildren()[i] = new
			}
		}
	}
	// Wire up old parent to new child
	new.SetParent(oldParent)
}

// Get nth random node out of the tree for random mutation
func GetNthNode(node Node, n, count int) (Node, int) {
	// We have arrived
	if n == count {
		return node, count
	}
	var result Node // Set up an empty node
	// Iterate through the node's children
	for _, child := range node.GetChildren() {
		count++
		result, count = GetNthNode(child, n, count)
		if result != nil {
			// We have arrived at the (n)th node
			return result, count
		}
	}
	return nil, count // We never found the nth node for this step in the iteration
}
