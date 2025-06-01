package merkleTree

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

// TestMerkleTree tests the Merkle Tree implementation
func TestMerkleTree(t *testing.T) {
	// Test case 1: Create a Merkle tree with multiple data blocks
	t.Run("Create Merkle Tree", func(t *testing.T) {
		// Create test data
		data := [][]byte{
			[]byte("Block 1"),
			[]byte("Block 2"),
			[]byte("Block 3"),
			[]byte("Block 4"),
		}

		// Create Merkle tree
		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree: %v", err)
		}

		// Verify tree structure
		if tree.RootNode == nil {
			t.Fatalf("Tree root node is nil")
		}
	})

	// Test case 2: Generate and verify a valid proof
	t.Run("Generate and Verify Valid Proof", func(t *testing.T) {
		// Create test data
		data := [][]byte{
			[]byte("Block 1"),
			[]byte("Block 2"),
			[]byte("Block 3"),
			[]byte("Block 4"),
		}

		// Create Merkle tree
		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree: %v", err)
		}

		// Generate proof for Block 2
		targetData := []byte("Block 2")
		proof, proofFlags, err := tree.GenerateProof(targetData)
		if err != nil {
			t.Fatalf("Failed to generate proof: %v", err)
		}

		// Calculate the hash of target data
		targetHash := sha256.Sum256(targetData)

		// Verify the proof
		valid, err := tree.VerifyProof(targetHash[:], proof, proofFlags)
		if err != nil {
			t.Fatalf("Failed to verify proof: %v", err)
		}

		if !valid {
			t.Errorf("Proof verification failed for valid data")
		}
	})

	// Test case 3: Verify invalid proof (wrong data)
	t.Run("Verify Invalid Proof - Wrong Data", func(t *testing.T) {
		// Create test data
		data := [][]byte{
			[]byte("Block 1"),
			[]byte("Block 2"),
			[]byte("Block 3"),
			[]byte("Block 4"),
		}

		// Create Merkle tree
		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree: %v", err)
		}

		// Generate proof for Block 2
		targetData := []byte("Block 2")
		proof, proofFlags, err := tree.GenerateProof(targetData)
		if err != nil {
			t.Fatalf("Failed to generate proof: %v", err)
		}

		// Try to verify with wrong data
		wrongData := []byte("Block X")
		wrongHash := sha256.Sum256(wrongData)

		valid, err := tree.VerifyProof(wrongHash[:], proof, proofFlags)
		if err != nil {
			t.Fatalf("Failed to verify proof: %v", err)
		}

		if valid {
			t.Errorf("Proof verification incorrectly succeeded for invalid data")
		}
	})

	// Test case 4: Verify invalid proof (tampered proof)
	t.Run("Verify Invalid Proof - Tampered Proof", func(t *testing.T) {
		// Create test data
		data := [][]byte{
			[]byte("Block 1"),
			[]byte("Block 2"),
			[]byte("Block 3"),
			[]byte("Block 4"),
		}

		// Create Merkle tree
		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree: %v", err)
		}

		// Generate proof for Block 2
		targetData := []byte("Block 2")
		proof, proofFlags, err := tree.GenerateProof(targetData)
		if err != nil {
			t.Fatalf("Failed to generate proof: %v", err)
		}

		// Tamper with the proof (if proof exists)
		if len(proof) > 0 {
			tamperedProof := make([][]byte, len(proof))
			copy(tamperedProof, proof)
			tamperedProof[0] = []byte("tampered data")

			targetHash := sha256.Sum256(targetData)
			valid, err := tree.VerifyProof(targetHash[:], tamperedProof, proofFlags)
			if err != nil {
				t.Fatalf("Failed to verify proof: %v", err)
			}

			if valid {
				t.Errorf("Proof verification incorrectly succeeded with tampered proof")
			}
		}
	})

	// Test case 5: Test edge case - verify data that exists in the tree
	t.Run("Verify Data Exists in Tree", func(t *testing.T) {
		// Create test data
		data := [][]byte{
			[]byte("Block 1"),
			[]byte("Block 2"),
			[]byte("Block 3"),
			[]byte("Block 4"),
		}

		// Create Merkle tree
		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree: %v", err)
		}

		// Verify existing data
		exists, err := tree.VerifyData([]byte("Block 3"))
		if err != nil {
			t.Fatalf("Failed to verify data: %v", err)
		}

		if !exists {
			t.Errorf("Failed to verify existing data in the tree")
		}

		// Verify non-existing data
		exists, err = tree.VerifyData([]byte("Block X"))
		if err != nil {
			t.Fatalf("Failed to verify data: %v", err)
		}

		if exists {
			t.Errorf("Incorrectly verified non-existing data in the tree")
		}
	})

	// Test case 6: Test edge case - empty data
	t.Run("Edge Case - Empty Data", func(t *testing.T) {
		// Attempt to create tree with empty data slice
		_, err := NewMerkleTree([][]byte{})
		if err == nil {
			t.Errorf("Expected error when creating tree with empty data, but got none")
		}

		// Create tree with a single empty block
		data := [][]byte{
			[]byte(""),
		}

		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree with single empty block: %v", err)
		}

		// Verify the empty data
		exists, err := tree.VerifyData([]byte(""))
		if err != nil {
			t.Fatalf("Failed to verify empty data: %v", err)
		}

		if !exists {
			t.Errorf("Failed to verify empty data in the tree")
		}
	})

	// Test case 7: Test edge case - odd number of data blocks
	t.Run("Edge Case - Odd Number of Blocks", func(t *testing.T) {
		// Create test data with odd number of blocks
		data := [][]byte{
			[]byte("Block 1"),
			[]byte("Block 2"),
			[]byte("Block 3"),
		}

		// Create Merkle tree
		tree, err := NewMerkleTree(data)
		if err != nil {
			t.Fatalf("Failed to create Merkle tree with odd number of blocks: %v", err)
		}

		// The last block should be duplicated, so both Block 3 should be verifiable
		exists, err := tree.VerifyData([]byte("Block 3"))
		if err != nil {
			t.Fatalf("Failed to verify data: %v", err)
		}

		if !exists {
			t.Errorf("Failed to verify data in tree with odd number of blocks")
		}
	})
}

// TestMerkleNode tests the Merkle Node creation
func TestMerkleNode(t *testing.T) {
	// Test creating a leaf node
	t.Run("Create Leaf Node", func(t *testing.T) {
		data := []byte("Leaf data")
		node, err := NewMerkleNode(nil, nil, data)
		if err != nil {
			t.Fatalf("Failed to create leaf node: %v", err)
		}

		expectedHash := sha256.Sum256(data)
		if !bytes.Equal(node.Data, expectedHash[:]) {
			t.Errorf("Leaf node hash incorrect. Expected: %x, Got: %x", expectedHash[:], node.Data)
		}
	})

	// Test creating an internal node
	t.Run("Create Internal Node", func(t *testing.T) {
		leftData := []byte("Left data")
		rightData := []byte("Right data")

		leftNode, err := NewMerkleNode(nil, nil, leftData)
		if err != nil {
			t.Fatalf("Failed to create left node: %v", err)
		}

		rightNode, err := NewMerkleNode(nil, nil, rightData)
		if err != nil {
			t.Fatalf("Failed to create right node: %v", err)
		}

		parentNode, err := NewMerkleNode(leftNode, rightNode, nil)
		if err != nil {
			t.Fatalf("Failed to create parent node: %v", err)
		}

		// Calculate expected hash
		combinedData := append(leftNode.Data, rightNode.Data...)
		expectedHash := sha256.Sum256(combinedData)

		if !bytes.Equal(parentNode.Data, expectedHash[:]) {
			t.Errorf("Parent node hash incorrect. Expected: %x, Got: %x", expectedHash[:], parentNode.Data)
		}
	})

	// Test error cases
	t.Run("Error Cases", func(t *testing.T) {
		// Test error when creating leaf node with nil data
		_, err := NewMerkleNode(nil, nil, nil)
		if err == nil {
			t.Errorf("Expected error when creating leaf node with nil data, but got none")
		}

		// Test error when creating internal node with missing children
		leftData := []byte("Left data")
		leftNode, _ := NewMerkleNode(nil, nil, leftData)

		_, err = NewMerkleNode(leftNode, nil, nil)
		if err == nil {
			t.Errorf("Expected error when creating internal node with missing right child, but got none")
		}

		_, err = NewMerkleNode(nil, leftNode, nil)
		if err == nil {
			t.Errorf("Expected error when creating internal node with missing left child, but got none")
		}
	})
}
