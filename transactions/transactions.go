package transactions

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/gob"
    "encoding/hex"
    "errors"
    "fmt"
    "log"
    "math/big"
    
    "github.com/omsingh/blockchain/wallet"
)

// Transaction represents a blockchain transaction
type Transaction struct {
    ID   []byte
    Vin  []TxInput
    Vout []TxOutput
}

// Block represents a block in the blockchain
type Block struct {
    Timestamp     int64
    Transactions  []*Transaction
    PrevBlockHash []byte
    Hash          []byte
    Nonce         int
}

// SignTransaction signs each input of a Transaction using the provided wallet
func (bc *Blockchain) SignTransaction(tx *Transaction, walletInstance *wallet.Wallet) error {
    prevTXs := make(map[string]Transaction)

    for _, vin := range tx.Vin {
        if len(vin.Txid) == 0 {
            continue // Skip coinbase
        }
        
        prevTX, err := bc.FindTransaction(vin.Txid)
        if err != nil {
            return fmt.Errorf("failed to find previous transaction: %v", err)
        }
        
        prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
    }

    return tx.Sign(walletInstance, prevTXs)
}

// VerifyTransaction verifies signature of Transaction inputs
func (bc *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
    if tx.IsCoinbase() {
        return true, nil
    }
    
    prevTXs := make(map[string]Transaction)

    for _, vin := range tx.Vin {
        if len(vin.Txid) == 0 {
            continue // Skip coinbase
        }
        
        prevTX, err := bc.FindTransaction(vin.Txid)
        if err != nil {
            return false, fmt.Errorf("failed to find previous transaction: %v", err)
        }
        
        prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
    }

    return tx.Verify(prevTXs)
}

// Blockchain represents a chain of blocks
type Blockchain struct {
    // Fields omitted for brevity
}

// FindTransaction finds a transaction by ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
    bci := bc.Iterator()

    for {
        block := bci.Next()

        for _, tx := range block.Transactions {
            if bytes.Compare(tx.ID, ID) == 0 {
                return *tx, nil
            }
        }

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }

    return Transaction{}, errors.New("transaction not found")
}

// TxInput represents a transaction input
type TxInput struct {
    Txid      []byte // The ID of the transaction containing the output to spend
    Vout      int    // The index of the output in the transaction
    Signature []byte // The digital signature that proves ownership
    PubKey    []byte // The public key of the sender
}

// TxOutput represents a transaction output
type TxOutput struct {
    Value      int    // The amount of coins
    PubKeyHash []byte // The hash of the public key (address) of the recipient
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
    var hash [32]byte
    txCopy := *tx
    txCopy.ID = []byte{}
    
    encoded, err := serializeTransaction(txCopy)
    if err != nil {
        log.Panic(err)
    }
    
    hash = sha256.Sum256(encoded)
    return hash[:]
}

// serializeTransaction serializes a transaction
func serializeTransaction(tx Transaction) ([]byte, error) {
    var encoded bytes.Buffer
    enc := gob.NewEncoder(&encoded)
    
    err := enc.Encode(tx)
    if err != nil {
        return nil, fmt.Errorf("failed to encode transaction: %v", err)
    }
    
    return encoded.Bytes(), nil
}

// IsCoinbase checks whether the transaction is coinbase
func (tx *Transaction) IsCoinbase() bool {
    return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
    var inputs []TxInput
    var outputs []TxOutput

    for _, vin := range tx.Vin {
        inputs = append(inputs, TxInput{vin.Txid, vin.Vout, nil, nil})
    }

    for _, vout := range tx.Vout {
        outputs = append(outputs, TxOutput{vout.Value, vout.PubKeyHash})
    }

    txCopy := Transaction{tx.ID, inputs, outputs}

    return txCopy
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(walletInstance *wallet.Wallet, prevTXs map[string]Transaction) error {
    if tx.IsCoinbase() {
        return nil
    }

    // Verify all referenced inputs exist in prevTXs
    for _, vin := range tx.Vin {
        if len(vin.Txid) == 0 {
            continue // Skip coinbase
        }
        
        txID := hex.EncodeToString(vin.Txid)
        _, exists := prevTXs[txID]
        if !exists {
            return fmt.Errorf("referenced input transaction not found: %s", txID)
        }
    }

    txCopy := tx.TrimmedCopy()

    for inID, vin := range txCopy.Vin {
        if len(vin.Txid) == 0 {
            continue // Skip coinbase
        }
        
        txID := hex.EncodeToString(vin.Txid)
        prevTx := prevTXs[txID]
        txCopy.Vin[inID].Signature = nil
        txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
        txCopy.ID = txCopy.Hash()
        txCopy.Vin[inID].PubKey = nil

        // Use wallet's SignData function for signing
        signature, err := walletInstance.SignData(txCopy.ID)
        if err != nil {
            return fmt.Errorf("failed to sign transaction input: %v", err)
        }
        
        tx.Vin[inID].Signature = signature
    }

    return nil
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
    if tx.IsCoinbase() {
        return true, nil
    }

    // Verify all referenced inputs exist in prevTXs
    for _, vin := range tx.Vin {
        if len(vin.Txid) == 0 {
            continue // Skip coinbase
        }
        
        txID := hex.EncodeToString(vin.Txid)
        _, exists := prevTXs[txID]
        if !exists {
            return false, fmt.Errorf("referenced input transaction not found: %s", txID)
        }
    }

    txCopy := tx.TrimmedCopy()

    for inID, vin := range tx.Vin {
        if len(vin.Txid) == 0 {
            continue // Skip coinbase
        }
        
        txID := hex.EncodeToString(vin.Txid)
        prevTx := prevTXs[txID]
        txCopy.Vin[inID].Signature = nil
        txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
        txCopy.ID = txCopy.Hash()
        txCopy.Vin[inID].PubKey = nil

        // Verify signature using wallet package function
        if !wallet.VerifySignature(vin.PubKey, txCopy.ID, vin.Signature) {
            return false, nil
        }
    }

    return true, nil
}

// ValidateTransaction validates a transaction
func (tx *Transaction) ValidateTransaction(prevTXs map[string]Transaction) error {
    if len(tx.ID) == 0 {
        return fmt.Errorf("transaction ID cannot be empty")
    }
    
    if len(tx.Vin) == 0 {
        return fmt.Errorf("transaction must have at least one input")
    }
    
    if len(tx.Vout) == 0 {
        return fmt.Errorf("transaction must have at least one output")
    }
    
    // Verify signatures if not a coinbase transaction
    if !tx.IsCoinbase() {
        valid, err := tx.Verify(prevTXs)
        if err != nil {
            return fmt.Errorf("signature verification error: %v", err)
        }
        if !valid {
            return fmt.Errorf("invalid transaction signature")
        }
    }
    
    return nil
}

// UsesKey checks whether the input uses the specified public key hash
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
    lockingHash := HashPubKey(in.PubKey)
    return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// IsLockedWithKey checks if the output is locked with the specified public key hash
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
    return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// BlockchainIterator represents an iterator over blockchain blocks
type BlockchainIterator struct {
    currentHash []byte
    blockchain  *Blockchain
}

// Iterator returns an iterator for the blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
    return &BlockchainIterator{bc.GetLastBlockHash(), bc}
}

// GetLastBlockHash returns the hash of the last block in the blockchain
func (bc *Blockchain) GetLastBlockHash() []byte {
    // Implementation would depend on how blockchain is stored
    // This is a placeholder
    return bc.lastHash
}

// Next returns the next block in the blockchain
func (i *BlockchainIterator) Next() *Block {
    var block *Block
    
    // Actual implementation would retrieve the block based on currentHash
    // This is a placeholder
    // block = GetBlockByHash(i.currentHash)
    
    if block != nil {
        i.currentHash = block.PrevBlockHash
    }
    
    return block
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
    // Verify all transactions
    for _, tx := range transactions {
        valid, err := bc.VerifyTransaction(tx)
        if err != nil {
            return nil, fmt.Errorf("transaction verification error: %v", err)
        }
        if !valid {
            return nil, errors.New("invalid transaction detected")
        }
    }
    
    lastHash := bc.GetLastBlockHash()
    
    // Create and mine new block
    newBlock := &Block{
        Timestamp:     time.Now().Unix(),
        Transactions:  transactions,
        PrevBlockHash: lastHash,
        Hash:          []byte{},
        Nonce:         0,
    }
    
    // ProofOfWork would be called here to mine the block
    // This is a placeholder
    // pow := NewProofOfWork(newBlock)
    // nonce, hash := pow.Run()
    // newBlock.Hash = hash
    // newBlock.Nonce = nonce
    
    // Add block to blockchain
    // bc.AddBlock(newBlock)
    
    return newBlock, nil
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(walletInstance *wallet.Wallet, to []byte, amount int, bc *Blockchain, findSpendableOutputs func([]byte, int) (int, map[string][]int, error)) (*Transaction, error) {
    var inputs []TxInput
    var outputs []TxOutput

    pubKeyHash := wallet.HashPubKey(walletInstance.PublicKey)

    // Find spendable outputs
    acc, validOutputs, err := findSpendableOutputs(pubKeyHash, amount)
    if err != nil {
        return nil, fmt.Errorf("failed to find spendable outputs: %v", err)
    }

    if acc < amount {
        return nil, fmt.Errorf("not enough funds: got %d, need %d", acc, amount)
    }

    // Build a list of inputs
    for txid, outs := range validOutputs {
        txID, err := hex.DecodeString(txid)
        if err != nil {
            return nil, fmt.Errorf("failed to decode transaction ID: %v", err)
        }

        for _, out := range outs {
            input := TxInput{
                Txid:      txID,
                Vout:      out,
                Signature: nil,
                PubKey:    walletInstance.PublicKey,
            }
            inputs = append(inputs, input)
        }
    }

    // Create the outputs
    toPubKeyHash := to
    outputs = append(outputs, TxOutput{
        Value:      amount,
        PubKeyHash: toPubKeyHash,
    })

    // If there is change, send it back to the sender
    if acc > amount {
        outputs = append(outputs, TxOutput{
            Value:      acc - amount,
            PubKeyHash: pubKeyHash,
        })
    }

    tx := &Transaction{
        ID:   []byte{},
        Vin:  inputs,
        Vout: outputs,
    }
    
    tx.ID = tx.Hash()
    
    // Sign the transaction
    if err := bc.SignTransaction(tx, walletInstance); err != nil {
        return nil, fmt.Errorf("failed to sign transaction: %v", err)
    }

    return tx, nil
}

// NewCoinbaseTx creates a new coinbase transaction
func NewCoinbaseTx(to []byte, data string) *Transaction {
    if data == "" {
        randData := make([]byte, 20)
        _, err := rand.Read(randData)
        if err != nil {
            log.Panic(err)
        }
        data = fmt.Sprintf("Reward to '%x'", randData)
    }

    txin := TxInput{
        Txid:      []byte{},
        Vout:      -1,
        Signature: nil,
        PubKey:    []byte(data),
    }

    txout := TxOutput{
        Value:      50, // Mining reward
        PubKeyHash: to,
    }

    tx := &Transaction{
        ID:   []byte{},
        Vin:  []TxInput{txin},
        Vout: []TxOutput{txout},
    }

    tx.ID = tx.Hash()

    return tx
}

