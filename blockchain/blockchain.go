// Define and manage chain aspect of the blockchain
package blockchain 

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	
	"github.com/OmSingh2003/blockchain-go/ProofOfWork"
	"github.com/OmSingh2003/blockchain-go/transactions"
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
	tip []byte      // Hash of the latest block
	db  *bbolt.DB   // Database connection
	mu  sync.RWMutex // Mutex for thread safety
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

// AddBlock creates and mines a new block with the given transactions.
// It validates the transactions, mines the block using proof of work,
// and adds it to the blockchain.
func (bc *Blockchain) AddBlock(txs []*transactions.Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	// Validate transactions
	for _, tx := range txs {
		if err := tx.ValidateTransaction(); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}

	// Ensure we have at least one transaction
	if len(txs) == 0 {
		// Add a coinbase transaction if none provided
		coinbase := NewCoinbaseTx("Miner", "Mining reward")
		txs = append(txs, coinbase)
	} else if !txs[0].IsCoinbase() {
		// Ensure the first transaction is a coinbase
		coinbase := NewCoinbaseTx("Miner", "Mining reward")
		txs = append([]*transactions.Transaction{coinbase}, txs...)
	}

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

	// Create new block
	newBlock := types.NewBlock(txs, lastHash)
	
	// Validate block before mining
	if err := newBlock.ValidateBlock(); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	// Mine the block
	ProofOfWork.MineBlock(newBlock)
	
	// Validate the proof of work
	pow := ProofOfWork.NewProofOfWork(newBlock)
	if !pow.Validate() {
		return errors.New("invalid block: proof of work validation failed")
	}
	
	// Store the new block in the database - use a single transaction to avoid race conditions
	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return errors.New("blocks bucket not found")
		}
		
		// Check for concurrent modifications
		currentTip := b.Get([]byte(lastHashKey))
		if !bytes.Equal(currentTip, lastHash) {
			return errors.New("blockchain tip has changed, please retry")
		}
		
		// Serialize the block
		blockData, err := newBlock.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize block: %v", err)
		}
		
		// Get the block hash
		blockHash := newBlock.CalculateHash()
		
		// Store the block
		if err := b.Put(blockHash, blockData); err != nil {
			return fmt.Errorf("failed to store block: %v", err)
		}
		
		// Update the last hash
		if err := b.Put([]byte(lastHashKey), blockHash); err != nil {
			return fmt.Errorf("failed to update last hash: %v", err)
		}
		
		// Update the tip
		bc.tip = blockHash
		
		return nil
	})
	
	return err
}

// NewCoinbaseTx creates a new coinbase transaction
func NewCoinbaseTx(to, data string) *transactions.Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := transactions.TxInput{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}

	txout := transactions.TxOutput{
		Value:        50, // Mining reward
		ScriptPubKey: to,
	}

	tx := &transactions.Transaction{
		ID:   []byte{},
		Vin:  []transactions.TxInput{txin},
		Vout: []transactions.TxOutput{txout},
	}

	// Set the transaction ID
	err := tx.SetID()
	if err != nil {
		fmt.Printf("Error setting transaction ID: %v\n", err)
	}

	return tx
}

// newGenesisBlock creates and returns the initial (genesis) block
func newGenesisBlock() *types.Block {
	coinbase := NewCoinbaseTx("Genesis", "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")
	block := types.NewBlock([]*transactions.Transaction{coinbase}, []byte{})
	ProofOfWork.MineBlock(block)
	return block
}

// NewBlockchain creates a new Blockchain or loads an existing one.
// If the blockchain database already exists, it opens it and returns
// the blockchain with the current tip. Otherwise, it creates a new
// blockchain with a genesis block.
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

		return &Blockchain{
			tip: tip,
			db:  db,
			mu:  sync.RWMutex{},
		}, nil
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
		
		// Validate the genesis block
		if err := genesis.ValidateBlock(); err != nil {
			return fmt.Errorf("invalid genesis block: %v", err)
		}
		
		blockData, err := genesis.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize genesis block: %v", err)
		}

		// Get the genesis block hash
		genesisHash := genesis.CalculateHash()

		// Store the block
		if err := b.Put(genesisHash, blockData); err != nil {
			return fmt.Errorf("failed to store genesis block: %v", err)
		}

		// Store the last hash
		if err := b.Put([]byte(lastHashKey), genesisHash); err != nil {
			return fmt.Errorf("failed to store last hash: %v", err)
		}

		tip = genesisHash
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize blockchain: %v", err)
	}

	return &Blockchain{
		tip: tip,
		db:  db,
		mu:  sync.RWMutex{},
	}, nil
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

// Iterator returns a BlockchainIterator for traversing the blockchain.
// It allows for iterating through all blocks in the blockchain in reverse
// order (from newest to oldest).
func (bc *Blockchain) Iterator() (*BlockchainIterator, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
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

// GetBlockByHash retrieves a block by its hash from the blockchain.
// It returns the block if found, or an error if not.
func (bc *Blockchain) GetBlockByHash(hash []byte) (*types.Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
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

// CloseDB safely closes the database connection.
// It should be called when the application exits to ensure
// all data is properly written to disk.
func (bc *Blockchain) CloseDB() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if bc.db != nil {
		return bc.db.Close()
	}
	return nil
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs for address
func (bc *Blockchain) FindUnspentTransactions(address string) ([]*transactions.Transaction, error) {
    bc.mu.RLock()
    defer bc.mu.RUnlock()

    var unspentTXs []*transactions.Transaction
    spentTXOs := make(map[string][]int)
    
    iterator, err := bc.Iterator()
    if err != nil {
        return nil, fmt.Errorf("failed to create blockchain iterator: %v", err)
    }

    for {
        block, err := iterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to get next block: %v", err)
        }
        if block == nil {
            break
        }

        for _, tx := range block.Transactions {
            txID := hex.EncodeToString(tx.ID)

        Outputs:
            for outIdx, out := range tx.Vout {
                if spentTXOs[txID] != nil {
                    for _, spentOut := range spentTXOs[txID] {
                        if spentOut == outIdx {
                            continue Outputs
                        }
                    }
                }

                if out.CanBeUnlockedWith(address) {
                    unspentTXs = append(unspentTXs, tx)
                }
            }

            if !tx.IsCoinbase() {
                for _, in := range tx.Vin {
                    if in.CanUnlockOutputWith(address) {
                        inTxID := hex.EncodeToString(in.Txid)
                        spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
                    }
                }
            }
        }

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }

    return unspentTXs, nil
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int, error) {
    bc.mu.RLock()
    defer bc.mu.RUnlock()

    unspentOutputs := make(map[string][]int)
    accumulated := 0

    unspentTXs, err := bc.FindUnspentTransactions(address)
    if err != nil {
        return 0, nil, fmt.Errorf("failed to find unspent transactions: %v", err)
    }

Work:
    for _, tx := range unspentTXs {
        txID := hex.EncodeToString(tx.ID)

        for outIdx, out := range tx.Vout {
            if out.CanBeUnlockedWith(address) && accumulated < amount {
                accumulated += out.Value
                unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

                if accumulated >= amount {
                    break Work
                }
            }
        }
    }

    if accumulated < amount {
        return accumulated, unspentOutputs, fmt.Errorf("not enough funds: got %d, need %d", accumulated, amount)
    }

    return accumulated, unspentOutputs, nil
}

// FindUTXO finds and returns all unspent transaction outputs for address
func (bc *Blockchain) FindUTXO(address string) ([]transactions.TxOutput, error) {
    var UTXOs []transactions.TxOutput
    
    unspentTransactions, err := bc.FindUnspentTransactions(address)
    if err != nil {
        return nil, fmt.Errorf("failed to find unspent transactions: %v", err)
    }

    for _, tx := range unspentTransactions {
        for _, out := range tx.Vout {
            if out.CanBeUnlockedWith(address) {
                UTXOs = append(UTXOs, out)
            }
        }
    }

    return UTXOs, nil
}
