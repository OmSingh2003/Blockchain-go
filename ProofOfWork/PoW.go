package ProofOfWork

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"

	"github.com/OmSingh2003/blockchain-go/types"
)

// targetBits defines the difficulty (e.g., number of leading zeros).
const targetBits = 24
const maxNonce = math.MaxInt64 // Max iterations for finding nonce.

// holds a block and the calculated difficulty target.
type ProofOfWork struct {
	block  *types.Block
	target *big.Int
}

//  creates a ProofOfWork instance with a calculated target.
func NewProofOfWork(b *types.Block) *ProofOfWork {
	target := big.NewInt(1)
	// Calculate target based on targetBits: target = 1 << (256 - targetBits)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{block: b, target: target}
	return pow
}

// prepareData creates the byte array to be hashed, including the nonce.
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	return pow.block.PrepareData(nonce, targetBits)
}

// Run performs the mining operation to find a valid nonce and hash.
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing data: \"%s\"\n", string(pow.block.Data))
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:]) // Convert hash to big.Int for comparison.

		// If hash is less than the target, a valid PoW is found.
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\rFound hash: %x (nonce %d)\n", hash, nonce)
			break
		}
		nonce++
		if nonce%200000 == 0 { 
			fmt.Printf("\rCurrent hash: %x (nonce %d)", hash, nonce)
		}
	}
	fmt.Print("\n") // Newline after loop, especially if feedback was printed.

	return nonce, hash[:]
}

// Validate checks if the block's Proof-of-Work is valid.
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce) // Use block's stored nonce.
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1 // Check if hash < target.
}

// MineBlock performs PoW for a block and updates its Hash and Nonce.
func MineBlock(block *types.Block) {
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce
}

