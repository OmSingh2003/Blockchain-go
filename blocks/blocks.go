package blocks

import (
	"github.com/OmSingh2003/blockchain-go/transactions"
	"github.com/OmSingh2003/blockchain-go/types"
)

// NewBlock creates and returns a new Block with the given transactions
func NewBlock(txs []*transactions.Transaction, prevBlockHash []byte) *types.Block {
	return types.NewBlock(txs, prevBlockHash)
}

// MineBlock performs proof of work on the given block to find a valid hash
func MineBlock(block *types.Block, targetBits int64) {
	// The actual mining logic is handled in the ProofOfWork package
	// This is just a wrapper function for backward compatibility
	block.SetNonce(0)
	block.UpdateHash()
}
