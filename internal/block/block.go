package block

import (
    "bytes"
    "crypto/sha256"
    "encoding/gob"
    "fmt"
    "strconv"
    "sync"
    "time"

    "github.com/OmSingh2003/blockchain-go/internal/transaction"
)

// Block represents a block in the blockchain
type Block struct {
    Timestamp     int64                        // Records when block was created/mined
    Transactions  []*transaction.Transaction   // stores Transactions 
    PrevBlockHash []byte                      // Stores the Hash of previous Block in the chain 
    Hash          []byte                      // Stores the Hash of current block in the chain
    Nonce         int                         // Number used in proof of work
    mu            sync.RWMutex                // Mutex for thread safety
}

// NewBlock creates and returns a new Block
func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte) *Block {
    block := &Block{
        Timestamp:     time.Now().Unix(),
        Transactions:  transactions,
        PrevBlockHash: prevBlockHash,
        Hash:          []byte{},
        Nonce:         0,
    }
    
    // Initialize the hash
    block.Hash = block.CalculateHash()
    return block
}

// IsGenesisBlock checks if this block is a genesis block
func (b *Block) IsGenesisBlock() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return len(b.PrevBlockHash) == 0
}

// HashTransactions returns a hash of the transactions in the block
// This is a thread-safe public method
func (b *Block) HashTransactions() []byte {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    return b.hashTransactionsInternal()
}

// hashTransactionsInternal is an internal method that doesn't use locks
// It should only be called when the lock is already held or when thread safety isn't required
func (b *Block) hashTransactionsInternal() []byte {
    var txHashes [][]byte
    
    for _, tx := range b.Transactions {
        txHashes = append(txHashes, tx.ID)
    }
    txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
    
    return txHash[:]
}

// PrepareData prepares data for hashing by concatenating block data with nonce
// This method is for internal use and doesn't acquire locks itself to prevent deadlocks
// Caller should ensure thread safety if needed
func (b *Block) PrepareData(nonce int, targetBits int64) []byte {
    data := bytes.Join(
        [][]byte{
            b.PrevBlockHash,
            b.hashTransactionsInternal(),
            IntToHex(b.Timestamp),
            IntToHex(targetBits),
            IntToHex(int64(nonce)),
        },
        []byte{},
    )
    return data
}

// Serialize serializes the block
func (b *Block) Serialize() ([]byte, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    
    err := encoder.Encode(b)
    if err != nil {
        return nil, err
    }
    
    return result.Bytes(), nil
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) (*Block, error) {
    var block Block
    
    decoder := gob.NewDecoder(bytes.NewReader(d))
    err := decoder.Decode(&block)
    if err != nil {
        return nil, err
    }
    
    return &block, nil
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
    return []byte(strconv.FormatInt(num, 10))
}

// ValidateBlock validates the block and its transactions
func (b *Block) ValidateBlock(prevTXs map[string]transaction.Transaction) error {
    b.mu.RLock()
    defer b.mu.RUnlock()

    // Special case for genesis block
    if len(b.PrevBlockHash) == 0 {
        if len(b.Transactions) != 1 || !b.Transactions[0].IsCoinbase() {
            return fmt.Errorf("invalid genesis block: must have exactly one coinbase transaction")
        }
        return nil
    }

    // Regular block validation
    if len(b.Transactions) == 0 {
        return fmt.Errorf("block must contain at least one transaction")
    }
    
    // Validate each transaction
    for i, tx := range b.Transactions {
        // Skip validation for coinbase transaction
        if tx.IsCoinbase() {
            continue
        }
        
        if err := tx.ValidateTransaction(prevTXs); err != nil {
            return fmt.Errorf("invalid transaction at index %d: %v", i, err)
        }
    }
    
    // Ensure first transaction is coinbase
    if !b.Transactions[0].IsCoinbase() {
        return fmt.Errorf("first transaction must be coinbase")
    }
    
    return nil
}

// UpdateHash updates the block's hash based on its current state
func (b *Block) UpdateHash() error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    // Calculate hash without locks since we already have the write lock
    data := b.PrepareData(b.Nonce, 24)
    hash := sha256.Sum256(data)
    b.Hash = hash[:]
    return nil
}

// CalculateHash calculates and returns the hash of the block
// This method acquires read lock to ensure thread safety
func (b *Block) CalculateHash() []byte {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    // Use 24 as the default target bits (same as in pow)
    data := b.PrepareData(b.Nonce, 24)
    hash := sha256.Sum256(data)
    return hash[:]
}

// SetNonce sets the nonce of the block in a thread-safe manner
func (b *Block) SetNonce(nonce int) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.Nonce = nonce
}

// GetHash returns the hash of the block in a thread-safe manner
func (b *Block) GetHash() []byte {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.Hash
}

// GetNonce returns the nonce of the block in a thread-safe manner
func (b *Block) GetNonce() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.Nonce
}
