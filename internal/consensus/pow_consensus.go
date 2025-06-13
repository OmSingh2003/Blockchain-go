package consensus

import (
	"fmt"
	"math/big"

	"github.com/OmSingh2003/decentralized-ledger/internal/block"
	"github.com/OmSingh2003/decentralized-ledger/internal/crypto/pow"
	"github.com/OmSingh2003/decentralized-ledger/internal/transaction"
	"github.com/OmSingh2003/decentralized-ledger/internal/wallet"
	"go.etcd.io/bbolt"
)

const (
	TARGET_BLOCK_TIME_SECONDS    = 600  // 10 minutes per block
	DIFFICULTY_ADJUSTMENT_BLOCKS = 2016 // Adjust difficulty every 2016 blocks
	MAX_ADJUSTMENT_FACTOR        = 4    // Limit difficulty change to 4x (1/4 or 4x)
	INITIAL_TARGET_BITS          = 24   // Starting difficulty for genesis block

	blocksBucket = "blocks" // Define blocksBucket constant here as well
)

// POWConsensus implements the consensus interface for POW
type POWConsensus struct {
	db *bbolt.DB // Referenced to blockchain database
	// NOTE: thinking of changing database
}

// NewPOWConsensus creates a new POWConsensus instance
func NewPOWConsensus(db *bbolt.DB) *POWConsensus {
	return &POWConsensus{db: db}
}

// Propose block for POW consensus is like finding a nonce
func (p *POWConsensus) ProposeBlock(proposerWallet *wallet.Wallet, transactions []*transaction.Transaction, prevBlockHash []byte, currentTipHash []byte) (*block.Block, error) {
	newBlock := block.NewBlock(transactions, prevBlockHash)

	// Determine targetBits for the new block
	currentTargetBits, err := p.getAdjustedTargetBits(currentTipHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get adjusted target bits: %v", err)
	}

	// Mine the block with the adjusted difficulty
	powInstance := pow.NewProofOfWork(newBlock, currentTargetBits)
	powInstance.Run() // This will also set the block's Nonce and Bits

	return newBlock, nil
}

// This block is for validating POW consensus
func (p *POWConsensus) ValidateBlock(b *block.Block, prevTXs map[string]transaction.Transaction) (bool, error) {
	// First, validate block structure and transactions (similar to existing block.ValidateBlock)
	if err := b.ValidateBlock(prevTXs); err != nil {
		return false, fmt.Errorf("block structure/transaction validation failed: %v", err)
	}

	// Then, validate Proof-of-Work
	powCheck := pow.NewProofOfWork(b, b.GetBits()) // Use block's stored bits for validation
	return powCheck.Validate(), nil
}

// Getting difficulty for POW returns the current targetBits
func (p *POWConsensus) GetCurrentDifficulty(blockchainTipHash []byte) (interface{}, error) {
	return p.getAdjustedTargetBits(blockchainTipHash)
}

// getAdjustedTargetBits calculates and returns the current targetBits for mining.
// This function is moved and adapted from blockchain.go
func (p *POWConsensus) getAdjustedTargetBits(currentTipHash []byte) (int64, error) {
	var currentBlock *block.Block
	var err error

	err = p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		blockData := b.Get(currentTipHash)
		if blockData == nil {
			return fmt.Errorf("tip block not found")
		}
		currentBlock, err = block.DeserializeBlock(blockData)
		return err
	})
	if err != nil {
		return 0, err
	}

	// For the genesis block, return the initial target bits
	if currentBlock.IsGenesisBlock() {
		return INITIAL_TARGET_BITS, nil
	}

	// To accurately get the height and previous blocks for adjustment,
	// we need to iterate backwards or store block height in the block.
	// For simplicity, this adaptation still iterates backwards.
	// In a production system, store block height in Block for efficiency.

	// Get the previous block (needed to determine its Bits for non-adjustment periods)
	prevBlock, err := p.findBlock(currentBlock.PrevBlockHash)
	if err != nil {
		return 0, fmt.Errorf("failed to find previous block for difficulty adjustment: %v", err)
	}

	// Get current block height (approximate, better to store in block)
	currentHeight := int64(0)
	tempHash := currentTipHash
	err = p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		for {
			blkData := b.Get(tempHash)
			if blkData == nil {
				break
			}
			blk, e := block.DeserializeBlock(blkData)
			if e != nil {
				return e
			}
			currentHeight++
			if len(blk.PrevBlockHash) == 0 { // Genesis block
				break
			}
			tempHash = blk.PrevBlockHash
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// Adjust difficulty only after a certain number of blocks
	if currentHeight > 0 && currentHeight%DIFFICULTY_ADJUSTMENT_BLOCKS == 0 {
		// Find the first block of the last adjustment period
		firstBlockOfPeriodHash := currentTipHash

		// Create a temporary iterator to go back DIFFICULTY_ADJUSTMENT_BLOCKS
		tempIteratorHash := currentTipHash
		for i := 0; i < DIFFICULTY_ADJUSTMENT_BLOCKS-1; i++ { // Go back (N-1) blocks
			var tempBlock *block.Block
			err := p.db.View(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(blocksBucket))
				blkData := b.Get(tempIteratorHash)
				if blkData == nil {
					return fmt.Errorf("block not found while iterating backwards")
				}
				tempBlock, err = block.DeserializeBlock(blkData)
				return err
			})
			if err != nil {
				return 0, err
			}
			tempIteratorHash = tempBlock.PrevBlockHash
			if len(tempIteratorHash) == 0 { // Reached genesis block before full interval
				break
			}
		}
		firstBlockOfPeriodHash = tempIteratorHash // This is the hash of the block at the start of the interval

		firstBlockOfPeriod, err := p.findBlock(firstBlockOfPeriodHash)
		if err != nil {
			return 0, fmt.Errorf("failed to find first block of adjustment period: %v", err)
		}

		actualTimeTaken := currentBlock.Timestamp - firstBlockOfPeriod.Timestamp
		expectedTimeTaken := int64(DIFFICULTY_ADJUSTMENT_BLOCKS) * TARGET_BLOCK_TIME_SECONDS

		currentTarget := big.NewInt(1)
		currentTarget.Lsh(currentTarget, uint(256-prevBlock.GetBits())) // Get target from previous block's bits

		// Calculate new target
		newTarget := new(big.Int).Set(currentTarget)
		newTarget.Mul(newTarget, big.NewInt(actualTimeTaken))
		newTarget.Div(newTarget, big.NewInt(expectedTimeTaken))

		// Apply limits to prevent extreme difficulty changes
		maxTarget := new(big.Int).Set(currentTarget)
		maxTarget.Mul(maxTarget, big.NewInt(MAX_ADJUSTMENT_FACTOR))

		minTarget := new(big.Int).Set(currentTarget)
		minTarget.Div(minTarget, big.NewInt(MAX_ADJUSTMENT_FACTOR))

		if newTarget.Cmp(maxTarget) == 1 { // if newTarget > maxTarget
			newTarget.Set(maxTarget)
		} else if newTarget.Cmp(minTarget) == -1 { // if newTarget < minTarget
			newTarget.Set(minTarget)
		}

		// Convert new target back to bits
		newTargetBits := int64(256 - newTarget.BitLen())
		if newTargetBits < 1 { // Ensure targetBits doesn't go below 1
			newTargetBits = 1
		}

		return newTargetBits, nil

	} else {
		// If not adjustment period, use the targetBits from the previous block
		return prevBlock.GetBits(), nil
	}
}

// findBlock is a helper function to fetch a block by its hash from the database.
func (p *POWConsensus) findBlock(hash []byte) (*block.Block, error) {
	var blockData []byte
	err := p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData = b.Get(hash)
		if blockData == nil {
			return fmt.Errorf("block not found for hash: %x", hash)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	blk, err := block.DeserializeBlock(blockData)
	if err != nil {
		return nil, err
	}
	return blk, nil
}
