package pow

import (
    "crypto/sha256"
    "fmt"
    "math"
    "math/big"

    "github.com/OmSingh2003/decentralized-ledger/internal/block"
)

// targetBits defines the difficulty (e.g., number of leading zeros).
const targetBits = 24
const maxNonce = math.MaxInt64 // Max iterations for finding nonce.

// ProofOfWork holds a block and the calculated difficulty target.
type ProofOfWork struct {
    block  *block.Block
    target *big.Int
}

// NewProofOfWork creates a new proof-of-work instance with the given block.
func NewProofOfWork(b *block.Block) *ProofOfWork {
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits))

    pow := &ProofOfWork{b, target}
    return pow
}

// Run performs the proof-of-work computation.
func (pow *ProofOfWork) Run() {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0

    fmt.Printf("Mining a new block")
    for nonce < maxNonce {
        data := pow.block.PrepareData(nonce, targetBits)
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
    pow.block.UpdateHash()
}

// Validate validates proof-of-work
func (pow *ProofOfWork) Validate() bool {
    var hashInt big.Int

    data := pow.block.PrepareData(pow.block.GetNonce(), targetBits)
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])

    isValid := hashInt.Cmp(pow.target) == -1

    return isValid
}
