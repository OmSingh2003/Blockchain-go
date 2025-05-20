// Define and manage chain aspect of the blockchain
package blockchain 

import (
	"github.com/OmSingh2003/blockchain-go/ProofOfWork"
	"github.com/OmSingh2003/blockchain-go/types"
)

type Blockchain struct {
	Blocks []*types.Block
}

// AddBlock creates and mines a new block with the given data
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := types.NewBlock(data, prevBlock.Hash)
	
	// Mine the block using ProofOfWork
	ProofOfWork.MineBlock(newBlock)
	
	bc.Blocks = append(bc.Blocks, newBlock)
}

// newGenesisBlock creates and returns the initial (genesis) block
func newGenesisBlock() *types.Block {
	block := types.NewBlock("Genesis Block", []byte{})
	ProofOfWork.MineBlock(block)
	return block
}

// NewBlockchain creates and returns a new Blockchain instance
func NewBlockchain() *Blockchain {
	return &Blockchain{Blocks: []*types.Block{newGenesisBlock()}}
}

// ValidateBlock validates the proof of work for a given block
func ValidateBlock(block *types.Block) bool {
	pow := ProofOfWork.NewProofOfWork(block)
	return pow.Validate()
}
