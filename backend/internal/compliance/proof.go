package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// MerkleNode represents a node in the merkle tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// MerkleTree represents the entire tree
type MerkleTree struct {
	RootNode *MerkleNode
	Leafs    [][]byte
	mu       sync.Mutex
}

// NewMerkleTree creates a new tree from a list of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []*MerkleNode

	if len(data) == 0 {
		return &MerkleTree{}
	}

	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, node)
	}

	// Build tree up
	for len(nodes) > 1 {
		var newLevel []*MerkleNode

		for i := 0; i < len(nodes); i += 2 {
			node1 := nodes[i]
			var node2 *MerkleNode

			if i+1 < len(nodes) {
				node2 = nodes[i+1]
			} else {
				// Duplicate last node if odd number
				node2 = nodes[i]
			}

			parentNode := NewMerkleNode(node1, node2, nil)
			newLevel = append(newLevel, parentNode)
		}

		nodes = newLevel
	}

	return &MerkleTree{
		RootNode: nodes[0],
		Leafs:    data,
	}
}

// NewMerkleNode creates a new node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := &MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return node
}

// VerifyProof verifies if a piece of data belongs to the tree
// For PoC, we just expose the Root Hash to show integrity changes
func (t *MerkleTree) GetRootHash() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.RootNode == nil {
		return ""
	}
	return hex.EncodeToString(t.RootNode.Data)
}

// AddLeaf adds a leaf and rebuilds the tree (Simplified, expensive)
func (t *MerkleTree) AddLeaf(data []byte) {
	t.mu.Lock()
	leave := append(t.Leafs, data)
	t.mu.Unlock()
	
	// Rebuild
	newTree := NewMerkleTree(leave)
	t.mu.Lock()
	t.RootNode = newTree.RootNode
	t.Leafs = leave
	t.mu.Unlock()
}
