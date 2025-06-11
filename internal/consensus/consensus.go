package consensus

import (
	"github.com/OmSingh2003/decentralized-ledger/internal/block"
	"github.com/OmSingh2003/decentralized-ledger/internal/blockchain"
	"github.com/OmSingh2003/decentralized-ledger/internal/transaction"
)

//consensus defines the inteface for different blockchain algorithms
type Consensus interface {
	//Propose block is responisble for creating a new block according to Consensus rule 
	// For POW this would involve finding a nonce. For POS , selecting a validator and signing.
	// it returns the newly created block or an error 
	ProposeBlock (transaction []*transaction.Transaction, prevBlockHash []byte, currentTipHash []byte) (*block.Block , error )
	//Validate Block checks if a given block is valid according to the Consensus rule
	//For POW,this involves validating the nonce and hash . For POS, validating signature and stake
	//It returns true if the block is valid , along with any error encountered during validating
	ValidateBlock(block *block.Block, prevTXs map[string]transaction.Transaction) (bool,error)
	//GetCurrentDifficulty returns the curernt difficulty / target information required for new block creation 
	// For POW , this would be the targetBits . For POS , it might be the current validator set
	GetCurrentDifficulty(blockchainTipHash []byte) (inteface {}, error)
}
