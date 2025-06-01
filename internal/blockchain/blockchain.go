package blockchain

import (
    "encoding/hex"
    "fmt"
    "os"
    "sync"

    "github.com/OmSingh2003/blockchain-go/internal/block"
    "github.com/OmSingh2003/blockchain-go/internal/crypto/pow"
    "github.com/OmSingh2003/blockchain-go/internal/transaction"
    "github.com/OmSingh2003/blockchain-go/internal/wallet"
    "go.etcd.io/bbolt"
)

const (
    dbFile              = "blockchain.db"
    blocksBucket        = "blocks"
    lastHashKey         = "l" // Key for storing the last block hash
    genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
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

// NewBlockchain opens an existing blockchain
func NewBlockchain() (*Blockchain, error) {
    // Only open existing blockchain
    if !DbExists() {
        return nil, fmt.Errorf("no existing blockchain found")
    }
    
    db, err := bbolt.Open(dbFile, 0600, nil)
    if err != nil {
        return nil, fmt.Errorf("cannot open blockchain db: %v", err)
    }

    var tip []byte
    err = db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        if b == nil {
            return fmt.Errorf("no existing blockchain found")
        }
        tip = b.Get([]byte(lastHashKey))
        return nil
    })
    if err != nil {
        db.Close()
        return nil, err
    }

    bc := Blockchain{tip, db, sync.RWMutex{}}
    return &bc, nil
}

// CreateBlockchain creates a new blockchain with a genesis block
func CreateBlockchain(minerWallet *wallet.Wallet) (*Blockchain, error) {
    // Check if blockchain already exists
    if DbExists() {
        return nil, fmt.Errorf("blockchain already exists")
    }
    
    // Validate miner wallet
    if minerWallet == nil {
        return nil, fmt.Errorf("miner wallet is required to create blockchain")
    }
    
    // Open database
    db, err := bbolt.Open(dbFile, 0600, nil)
    if err != nil {
        return nil, fmt.Errorf("cannot open blockchain db: %v", err)
    }

    var tip []byte
    err = db.Update(func(tx *bbolt.Tx) error {
        // Create coinbase transaction with miner's address
        // Pass the public key directly - the hash will be calculated inside NewCoinbaseTx
        cbtx := transaction.NewCoinbaseTx(minerWallet.PublicKey, genesisCoinbaseData)
        genesis := block.NewBlock([]*transaction.Transaction{cbtx}, []byte{})

        // Mine the genesis block
        powInstance := pow.NewProofOfWork(genesis)
        powInstance.Run()

        // Create blocks bucket
        b, err := tx.CreateBucket([]byte(blocksBucket))
        if err != nil {
            return err
        }

        // Store the genesis block
        blockData, err := genesis.Serialize()
        if err != nil {
            return err
        }

        err = b.Put(genesis.Hash, blockData)
        if err != nil {
            return err
        }

        // Store the last block hash
        err = b.Put([]byte(lastHashKey), genesis.Hash)
        if err != nil {
            return err
        }

        tip = genesis.Hash
        return nil
    })
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to create genesis block: %v", err)
    }

    // Create blockchain instance
    bc := Blockchain{tip, db, sync.RWMutex{}}

    // Initialize UTXO set
    utxo := UTXOSet{&bc}
    err = utxo.Reindex()
    if err != nil {
        bc.CloseDB()
        return nil, fmt.Errorf("failed to initialize UTXO set: %v", err)
    }

    return &bc, nil
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) (*block.Block, error) {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    for _, tx := range transactions {
        if !tx.IsCoinbase() {
            if err := bc.VerifyTransaction(tx); err != nil {
                return nil, fmt.Errorf("invalid transaction: %v", err)
            }
        }
    }

    lastHash := bc.tip
    newBlock := block.NewBlock(transactions, lastHash)

    // Mine the block
    powInstance := pow.NewProofOfWork(newBlock)
    powInstance.Run()

    err := bc.db.Update(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        blockData, err := newBlock.Serialize()
        if err != nil {
            return err
        }

        err = b.Put(newBlock.Hash, blockData)
        if err != nil {
            return err
        }

        err = b.Put([]byte(lastHashKey), newBlock.Hash)
        if err != nil {
            return err
        }

        bc.tip = newBlock.Hash
        return nil
    })

    return newBlock, err
}

// Iterator returns a BlockchainIterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
    bc.mu.RLock()
    defer bc.mu.RUnlock()
    return &BlockchainIterator{bc.tip, bc.db}
}

// Next returns the next block from the iterator
func (i *BlockchainIterator) Next() (*block.Block, error) {
    var blockData []byte

    err := i.db.View(func(tx *bbolt.Tx) error {
        b := tx.Bucket([]byte(blocksBucket))
        blockData = b.Get(i.currentHash)
        return nil
    })
    if err != nil {
        return nil, err
    }

    if blockData == nil {
        return nil, nil
    }

    block, err := block.DeserializeBlock(blockData)
    if err != nil {
        return nil, err
    }

    i.currentHash = block.PrevBlockHash
    return block, nil
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string][]transaction.TxOutput {
    bc.mu.RLock()
    defer bc.mu.RUnlock()
    
    UTXO := make(map[string][]transaction.TxOutput)
    spentTXOs := make(map[string][]int)
    bci := bc.Iterator()

    for {
        block, err := bci.Next()
        if err != nil {
            break
        }
        if block == nil {
            break
        }

        for _, tx := range block.Transactions {
            txID := hex.EncodeToString(tx.ID)

        Outputs:
            for outIdx, out := range tx.Vout {
                // Was the output spent?
                if spentTXOs[txID] != nil {
                    for _, spentOutIdx := range spentTXOs[txID] {
                        if spentOutIdx == outIdx {
                            continue Outputs
                        }
                    }
                }

                outs := UTXO[txID]
                outs = append(outs, out)
                UTXO[txID] = outs
            }

            if !tx.IsCoinbase() {
                for _, in := range tx.Vin {
                    inTxID := hex.EncodeToString(in.Txid)
                    spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
                }
            }
        }

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }

    return UTXO
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) error {
    if tx.IsCoinbase() {
        return nil
    }

    prevTXs := make(map[string]transaction.Transaction)

    for _, vin := range tx.Vin {
        prevTX, err := bc.FindTransaction(vin.Txid)
        if err != nil {
            return err
        }
        if prevTX == nil {
            return fmt.Errorf("referenced transaction not found: %x", vin.Txid)
        }
        prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
    }

    valid, err := tx.Verify(prevTXs)
    if err != nil {
        return err
    }
    if !valid {
        return fmt.Errorf("invalid transaction signature")
    }

    return nil
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (*transaction.Transaction, error) {
    bci := bc.Iterator()

    for {
        block, err := bci.Next()
        if err != nil {
            return nil, err
        }
        if block == nil {
            break
        }

        for _, tx := range block.Transactions {
            if hex.EncodeToString(tx.ID) == hex.EncodeToString(ID) {
                return tx, nil
            }
        }

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }

    return nil, fmt.Errorf("transaction not found")
}

// CloseDB closes the database
func (bc *Blockchain) CloseDB() error {
    bc.mu.Lock()
    defer bc.mu.Unlock()
    return bc.db.Close()
}

// DbExists checks if the blockchain database exists
func DbExists() bool {
    _, err := os.Stat(dbFile)
    return !os.IsNotExist(err)
}
