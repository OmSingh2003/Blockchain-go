// Define and manage chain aspect of the blockchain
package blockchain 

import (
	"errors"
	"fmt"
	"os"
	
	"github.com/OmSingh2003/blockchain-go/ProofOfWork"
	"github.com/OmSingh2003/blockchain-go/types"
	"go.etcd.io/bbolt"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
	lastHashKey  = "l" // Key for storing the last block hash
)

// Blockchain represents the blockchain structure
type Blockchain struct {
	tip []byte   // Hash of the latest block
	db  *bbolt.DB // Database connection
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

// AddBlock creates and mines a new block with the given data
func (bc *Blockchain) AddBlock(data string) error {
	var lastHash []byte

	// Get the last block hash
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return errors.New("blocks bucket not found")
		}
		lastHash = b.Get([]byte(lastHashKey))
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get last block hash: %v", err)
	}

	// Create and mine the new block
	newBlock := types.NewBlock(data, lastHash)
	ProofOfWork.MineBlock(newBlock)
	
	// Validate the block
	if !ValidateBlock(newBlock) {
		return errors.New("invalid block: proof of work validation failed")
	}
	
	// Store the new block in the database
	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return errors.New("blocks bucket not found")
		}
		
		// Serialize the block
		blockData, err := newBlock.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize block: %v", err)
		}
		
		// Store the block
		err = b.Put(newBlock.Hash, blockData)
		if err != nil {
			return fmt.Errorf("failed to store block: %v", err)
		}
		
		// Update the last hash
		err = b.Put([]byte(lastHashKey), newBlock.Hash)
		if err != nil {
			return fmt.Errorf("failed to update last hash: %v", err)
		}
		
		// Update the tip
		bc.tip = newBlock.Hash
		
		return nil
	})
	
	return err
}

// newGenesisBlock creates and returns the initial (genesis) block
func newGenesisBlock() *types.Block {
	block := types.NewBlock("Genesis Block", []byte{})
	ProofOfWork.MineBlock(block)
	return block
}

// NewBlockchain creates a new Blockchain or loads an existing one
func NewBlockchain() (*Blockchain, error) {
	// Check if the blockchain database already exists
	if dbExists() {
		fmt.Println("Blockchain already exists. Opening existing blockchain.")
		db, err := bbolt.Open(dbFile, 0600, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to open blockchain database: %v", err)
		}

		// Get the last block hash (tip)
		var tip []byte
		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			if b == nil {
				return errors.New("blocks bucket not found")
			}
			tip = b.Get([]byte(lastHashKey))
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get blockchain tip: %v", err)
		}

		return &Blockchain{tip, db}, nil
	}

	// Create a new blockchain with genesis block
	fmt.Println("No existing blockchain found. Creating a new one.")
	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blockchain database: %v", err)
	}

	var tip []byte
	err = db.Update(func(tx *bbolt.Tx) error {
		// Create the blocks bucket
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return fmt.Errorf("failed to create blocks bucket: %v", err)
		}

		// Create and store the genesis block
		genesis := newGenesisBlock()
		blockData, err := genesis.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize genesis block: %v", err)
		}

		err = b.Put(genesis.Hash, blockData)
		if err != nil {
			return fmt.Errorf("failed to store genesis block: %v", err)
		}

		// Store the last hash
		err = b.Put([]byte(lastHashKey), genesis.Hash)
		if err != nil {
			return fmt.Errorf("failed to store last hash: %v", err)
		}

		tip = genesis.Hash
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize blockchain: %v", err)
	}

	return &Blockchain{tip, db}, nil
}

// dbExists checks if the blockchain database already exists
func dbExists() bool {
	_, err := os.Stat(dbFile)
	return !os.IsNotExist(err)
}

// ValidateBlock validates the proof of work for a given block
func ValidateBlock(block *types.Block) bool {
	pow := ProofOfWork.NewProofOfWork(block)
	return pow.Validate()
}

// Iterator returns a BlockchainIterator for traversing the blockchain
func (bc *Blockchain) Iterator() (*BlockchainIterator, error) {
	if bc.tip == nil {
		return nil, errors.New("blockchain tip is nil")
	}
	return &BlockchainIterator{bc.tip, bc.db}, nil
}

// Next returns the next block from the blockchain
func (i *BlockchainIterator) Next() (*types.Block, error) {
	var block *types.Block

	// Get the block from the database
	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return errors.New("blocks bucket not found")
		}
		
		// Get the block data
		encodedBlock := b.Get(i.currentHash)
		if encodedBlock == nil {
			return nil // End of blockchain
		}
		
		// Deserialize the block
		var err error
		block, err = types.DeserializeBlock(encodedBlock)
		if err != nil {
			return fmt.Errorf("failed to deserialize block: %v", err)
		}
		
		// Move to the previous block
		i.currentHash = block.PrevBlockHash
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return block, nil
}

// GetBlockByHash retrieves a block by its hash
func (bc *Blockchain) GetBlockByHash(hash []byte) (*types.Block, error) {
	var block *types.Block

	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return errors.New("blocks bucket not found")
		}
		
		blockData := b.Get(hash)
		if blockData == nil {
			return fmt.Errorf("block %x not found", hash)
		}
		
		var err error
		block, err = types.DeserializeBlock(blockData)
		if err != nil {
			return fmt.Errorf("failed to deserialize block: %v", err)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return block, nil
}

// CloseDB closes the database connection
func (bc *Blockchain) CloseDB() error {
	if bc.db != nil {
		return bc.db.Close()
	}
	return nil
}
