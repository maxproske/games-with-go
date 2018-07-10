package game

// Pack our tree into contiguous memory
type priorityPos struct {
	Pos
	priority int
}

// Parent:		(i-1)/2
// Left child:	i*2+1
// Right child: i*2+2

// Implement go sort interface
type pqueue []priorityPos

// Return the new slice everytime we push
func (pq pqueue) push(pos Pos, priority int) pqueue {
	newNode := priorityPos{pos, priority}
	pq = append(pq, newNode) // Put at end of array to add to the heap (by shape property, add as right child)
	// Swap 5 and 8 if [5R,2L,8R]
	newNodeIndex := len(pq) - 1
	parentIndex, parent := pq.parent(newNodeIndex)
	// Because we are tracking cost, priority is lowest value
	for newNode.priority < parent.priority && newNodeIndex != 0 {
		pq.swap(newNodeIndex, parentIndex)
		newNodeIndex = parentIndex
		parentIndex, parent = pq.parent(newNodeIndex)
	}
	return pq
}

// Remove something from the queue
func (pq pqueue) pop() (pqueue, Pos) {
	result := pq[0].Pos // Always return the root
	// Replace root with rightmost leaf node
	pq[0] = pq[len(pq)-1]
	pq = pq[:len(pq)-1] // up to but not rightmode leaf node. shrink slice by 1

	// Edge case if length of pq is 0
	if len(pq) == 0 {
		return pq, result
	}

	// New root at 0
	index := 0
	node := pq[index]
	// New root may not be in the right place
	leftExists, leftIndex, left := pq.left(index)
	rightExists, rightIndex, right := pq.right(index)

	for (leftExists && node.priority > left.priority) ||
		(rightExists && node.priority > right.priority) {
		// Our node is not in the right spot yet, loop down the tree
		if !rightExists || left.priority <= right.priority {
			pq.swap(index, leftIndex)
			index = leftIndex
		} else {
			pq.swap(index, rightIndex)
			index = rightIndex
		}
		// Get information about new children
		leftExists, leftIndex, left = pq.left(index)
		rightExists, rightIndex, right = pq.right(index)
	}
	// Once the loop exists, we know we have a valid heap again
	return pq, result
}

func (pq pqueue) swap(i, j int) {
	tmp := pq[i]
	pq[i] = pq[j]
	pq[j] = tmp
}

func (pq pqueue) parent(i int) (int, priorityPos) {
	index := (i - 1) / 2
	return index, pq[index]
}

// Helper function to get left children
func (pq pqueue) left(i int) (bool, int, priorityPos) {
	index := i*2 + 1
	if index < len(pq) {
		// We have a child
		return true, index, pq[index]
	}
	return false, 0, priorityPos{}
}

// Helper function to get right children
func (pq pqueue) right(i int) (bool, int, priorityPos) {
	index := i*2 + 2
	if index < len(pq) {
		// We have a child
		return true, index, pq[index]
	}
	return false, 0, priorityPos{}
}
