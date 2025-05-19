// Define and manage chain aspect of the blockchain
package blockchain 

import (
	"github.com/OmSingh2003/blockchain-go/blocks"
)

type Blockchain struct { // struct which takes a field Blocks of type [] of  blocks which is Blocks dir 
	Blocks []*blocks.Block
}
// Appending a new block 
func (bc *Blockchain) AddBlock(data string) { 
	prevBlock := bc.Blocks[len(bc.Blocks)-1] // Accesses the last element of the `bc.Blocks` slice
	newBlock := blocks.NewBlock(data, prevBlock.Hash) //Calls the `NewBlock` function, which is imported from the `blocks` package.
	bc.Blocks = append(bc.Blocks, newBlock) // It adds `newBlock` to the end of the `bc.Blocks` slice
}

// newGenesisBlock is an unexported helper that creates and returns the initial (genesis) block.
// This block has a predefined data string and an empty previous hash, as it's the first in the chain.
func newGenesisBlock() *blocks.Block {
	return blocks.NewBlock("Genesis Block", []byte{})
}

// NewBlockchain is an exported constructor that creates and returns a new Blockchain instance.
// The new blockchain is always initialized with a genesis block.
func NewBlockchain() *Blockchain {
	// Initialize the Blockchain with its 'Blocks' slice containing the genesis block.
	return &Blockchain{Blocks: []*blocks.Block{newGenesisBlock()}}
}
