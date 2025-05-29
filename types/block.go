package types

import (
    "bytes"
    "crypto/sha256"
    "encoding/gob"
    "fmt"
    "strconv"
    "sync"
    "time"
)

// Block represents a block in the blockchain
type Block struct {
    Timestamp     int64          // Records when block was created/mined
    Transactions  []*Transaction // stores Transactions 
    PrevBlockHash []byte         // Stores the Hash of previous Block in the chain 
    Hash          []byte         // Stores the Hash of current block in the chain
    Nonce         int            // Number used in proof of work
    mu            sync.RWMutex   // Mutex for thread safety
}

// Transaction represents a blockchain transaction
type Transaction struct {
    ID   []byte
    Vin  []TxInput
    Vout []TxOutput
}

// TxInput represents a transaction input
type TxInput struct {
    Txid      []byte // The ID of the transaction containing the output to spend
    Vout      int    // The index of the output in the transaction
    ScriptSig string // The script that provides proof for spending
}

// TxOutput represents a transaction output
type TxOutput struct {
    Value        int    // The amount of coins
    ScriptPubKey string // The script that specifies spending conditions
}

// NewBlock creates and returns a new Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
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
func (b *Block) HashTransactions() []byte {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    var txHashes [][]byte
    
    for _, tx := range b.Transactions {
        txHashes = append(txHashes, tx.ID)
    }
    txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
    
    return txHash[:]
}

// PrepareData prepares data for hashing by concatenating block data with nonce
func (b *Block) PrepareData(nonce int, targetBits int64) []byte {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    data := bytes.Join(
        [][]byte{
            b.PrevBlockHash,
            b.HashTransactions(),
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

// SetID sets ID of a transaction
func (tx *Transaction) SetID() error {
    var encoded bytes.Buffer
    enc := gob.NewEncoder(&encoded)
    
    err := enc.Encode(tx)
    if err != nil {
        return err
    }
    
    hash := sha256.Sum256(encoded.Bytes())
    tx.ID = hash[:]
    
    return nil
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
    return []byte(strconv.FormatInt(num, 10))
}

// ValidateTransaction validates a transaction
func (tx *Transaction) ValidateTransaction() error {
    if len(tx.ID) == 0 {
        return fmt.Errorf("transaction ID cannot be empty")
    }
    
    if len(tx.Vin) == 0 {
        return fmt.Errorf("transaction must have at least one input")
    }
    
    if len(tx.Vout) == 0 {
        return fmt.Errorf("transaction must have at least one output")
    }
    
    return nil
}

// IsCoinbase checks whether the transaction is coinbase
func (tx *Transaction) IsCoinbase() bool {
    return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// ValidateBlock performs basic validation of the block
func (b *Block) ValidateBlock() error {
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
        if err := tx.ValidateTransaction(); err != nil {
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
    
    b.Hash = b.CalculateHash()
    return nil
}

// CalculateHash calculates and returns the hash of the block
func (b *Block) CalculateHash() []byte {
    // Don't lock here as this method is called from other methods that already have locks
    // Use 24 as the default target bits (same as in ProofOfWork)
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

// SetHash sets the hash of the block in a thread-safe manner
func (b *Block) SetHash(hash []byte) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.Hash = hash
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

