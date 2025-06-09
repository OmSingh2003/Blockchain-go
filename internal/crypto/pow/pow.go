package pow

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"

	"github.com/OmSingh2003/decentralized-ledger/internal/block"
)

// targetBits defines the difficulty (e.g., number of leading zeros).
// const targetBits = 24
const maxNonce = math.MaxInt64 // Max iterations for finding nonce.

// ProofOfWork holds a block and the calculated difficulty target.
type ProofOfWork struct {
	block      *block.Block
	target     *big.Int
	targetBits int64
}

// NewProofOfWork creates a new proof-of-work instance with the given block.
func NewProofOfWork(b *block.Block, targetBits int64) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target, targetBits} // stores targetBits
	return pow
}

// Run performs the proof-of-work computation.
func (pow *ProofOfWork) Run() {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		data := pow.block.PrepareData(nonce, pow.targetBits)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	pow.block.SetNonce(nonce)
	pow.block.UpdateHash()            // UpdateHash also needs targetBits,so i have to make sure it is handled properly
	pow.block.SetBits(pow.targetBits) // Set the bits in the block after mining
}

// Validate validates proof-of-work
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	// Get targetBits from block itself for validation
	data := pow.block.PrepareData(pow.block.GetNonce(), pow.targetBits)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

// GetTarget retuns the calculated difficult target
func (pow *ProofOfWork) GetTarget() *big.Int {
	return pow.target
}
