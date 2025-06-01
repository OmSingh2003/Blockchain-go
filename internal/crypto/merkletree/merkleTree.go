// Package merkleTree implements a Merkle Tree data structure for blockchain
// which provides an efficient way to verify the integrity of large datasets
// by using a tree of cryptographic hashes.
package merkleTree

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
)

// MerkleTree represents a Merkle tree with a reference to its root node.
type MerkleTree struct {
	RootNode *MerkleNode
}

// MerkleNode represents a node in the Merkle tree.
// Each node contains data (hash) and references to left and right child nodes.
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleNode creates a new Merkle node with the given left and right children,
// or as a leaf node with the hash of the provided data.
//
// If left and right are nil, it creates a leaf node with the hash of data.
// Otherwise, it creates an internal node by hashing the concatenation of left and right node hashes.
func NewMerkleNode(left, right *MerkleNode, data []byte) (*MerkleNode, error) {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		if data == nil {
			return nil, errors.New("cannot create leaf node with nil data")
		}
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		if left == nil || right == nil {
			return nil, errors.New("internal nodes must have both left and right children")
		}
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode, nil
}

// NewMerkleTree creates a new Merkle tree from a slice of data blocks.
// If the number of data blocks is odd, the last block is duplicated.
//
// Returns an error if data is empty or if there's an issue creating nodes.
func NewMerkleTree(data [][]byte) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot create a Merkle tree with no data")
	}

	// Create a copy of the data to avoid modifying the original slice
	dataBlocks := make([][]byte, len(data))
	copy(dataBlocks, data)

	// If there is an odd number of data blocks, duplicate the last one
	if len(dataBlocks)%2 != 0 {
		dataBlocks = append(dataBlocks, dataBlocks[len(dataBlocks)-1])
	}

	// Create leaf nodes
	var nodes []*MerkleNode
	for _, datum := range dataBlocks {
		node, err := NewMerkleNode(nil, nil, datum)
		if err != nil {
			return nil, fmt.Errorf("failed to create leaf node: %v", err)
		}
		nodes = append(nodes, node)
	}

	// Build the tree bottom-up
	for len(nodes) > 1 {
		var levelUp []*MerkleNode

		for i := 0; i < len(nodes); i += 2 {
			node, err := NewMerkleNode(nodes[i], nodes[i+1], nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create internal node: %v", err)
			}
			levelUp = append(levelUp, node)
		}

		nodes = levelUp
	}

	// The root is the only node left
	mTree := &MerkleTree{RootNode: nodes[0]}
	return mTree, nil
}

// VerifyData verifies if the given data is part of the Merkle tree
// by calculating its hash and checking if it exists in the tree.
func (m *MerkleTree) VerifyData(data []byte) (bool, error) {
	if m.RootNode == nil {
		return false, errors.New("merkle tree has no root")
	}

	// Calculate the hash of the data
	hash := sha256.Sum256(data)
	dataHash := hash[:]

	// Find the leaf node with matching data hash
	return m.findAndVerifyNode(m.RootNode, dataHash), nil
}

// findAndVerifyNode recursively searches for a node with the given hash
// in the subtree rooted at the current node.
func (m *MerkleTree) findAndVerifyNode(node *MerkleNode, hash []byte) bool {
	// Base case: leaf node
	if node.Left == nil && node.Right == nil {
		return bytes.Equal(node.Data, hash)
	}

	// Recursive case: internal node
	if node.Left != nil && m.findAndVerifyNode(node.Left, hash) {
		return true
	}
	if node.Right != nil && m.findAndVerifyNode(node.Right, hash) {
		return true
	}

	return false
}

// VerifyProof verifies a Merkle proof for a given data item.
// A Merkle proof consists of the minimal set of hashes needed to
// reconstruct the path from the data's leaf node to the root.
func (m *MerkleTree) VerifyProof(dataHash []byte, proof [][]byte, proofFlags []bool) (bool, error) {
	if m.RootNode == nil {
		return false, errors.New("merkle tree has no root")
	}

	calculatedHash := dataHash
	
	for i, hash := range proof {
		if proofFlags[i] {
			// Hash is on the right
			calculatedHash = hashPair(calculatedHash, hash)
		} else {
			// Hash is on the left
			calculatedHash = hashPair(hash, calculatedHash)
		}
	}
	
	return bytes.Equal(calculatedHash, m.RootNode.Data), nil
}

// GenerateProof generates a Merkle proof for a given data item.
// Returns the proof (array of hashes) and proof flags (indicating left/right position).
func (m *MerkleTree) GenerateProof(data []byte) ([][]byte, []bool, error) {
	if m.RootNode == nil {
		return nil, nil, errors.New("merkle tree has no root")
	}

	hash := sha256.Sum256(data)
	dataHash := hash[:]
	
	var proof [][]byte
	var proofFlags []bool
	
	// Find the path from root to the leaf containing dataHash
	if !m.collectProof(m.RootNode, dataHash, &proof, &proofFlags) {
		return nil, nil, errors.New("data not found in the merkle tree")
	}
	
	return proof, proofFlags, nil
}

// collectProof recursively collects hashes needed for the proof.
func (m *MerkleTree) collectProof(node *MerkleNode, targetHash []byte, proof *[][]byte, flags *[]bool) bool {
	// Base case: leaf node
	if node.Left == nil && node.Right == nil {
		return bytes.Equal(node.Data, targetHash)
	}
	
	// Check left subtree
	if node.Left != nil && m.collectProof(node.Left, targetHash, proof, flags) {
		// Add right sibling to proof
		*proof = append(*proof, node.Right.Data)
		*flags = append(*flags, true) // right position
		return true
	}
	
	// Check right subtree
	if node.Right != nil && m.collectProof(node.Right, targetHash, proof, flags) {
		// Add left sibling to proof
		*proof = append(*proof, node.Left.Data)
		*flags = append(*flags, false) // left position
		return true
	}
	
	return false
}

// GetRoot returns the root hash of the Merkle tree.
func (m *MerkleTree) GetRoot() []byte {
	if m.RootNode == nil {
		return nil
	}
	return m.RootNode.Data
}

// hashPair concatenates two hashes and returns their combined hash.
func hashPair(left, right []byte) []byte {
	combined := append(left, right...)
	hash := sha256.Sum256(combined)
	return hash[:]
}

// PrintTree prints the Merkle tree structure for debugging.
func (m *MerkleTree) PrintTree() {
	if m.RootNode == nil {
		fmt.Println("Empty tree")
		return
	}
	
	printNode(m.RootNode, 0)
}

// printNode recursively prints nodes with indentation based on depth.
func printNode(node *MerkleNode, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%sNode: %x\n", indent, node.Data)
	
	if node.Left != nil {
		fmt.Printf("%sLeft:\n", indent)
		printNode(node.Left, depth+1)
	}
	
	if node.Right != nil {
		fmt.Printf("%sRight:\n", indent)
		printNode(node.Right, depth+1)
	}
}

