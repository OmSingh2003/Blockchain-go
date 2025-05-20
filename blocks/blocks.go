package blocks
import (
	"bytes"
	"crypto/sha256"
	
	"github.com/OmSingh2003/blockchain-go/types"
)
// SetHash computes unique cryptographic hash for the Block 'b' and stores it in b.Hash
func SetHash(b *types.Block) {
	timestamp := types.IntToHex(b.Timestamp)
	// Prepare data for hashing
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp},[]byte{})
	hash := sha256.Sum256(headers) // Implement SHA256 as hashing algorithm 
	b.Hash = hash[:] // Assign hash value to the block
}
func NewBlock(data string, prevBlockHash []byte) *types.Block {
	// Create a new block using the types.NewBlock function
	block := types.NewBlock(data, prevBlockHash)
	
	
	return block
}

// MineBlock performs proof of work on the given block to find a valid hash
// This function will be called from the blockchain package
func MineBlock(block *types.Block, targetBits int64) {
	// This function is a placeholder to show how blocks will be mined
	SetHash(block)
}

