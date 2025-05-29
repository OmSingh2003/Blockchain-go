package transactions

import (
    "bytes"
    "crypto/sha256"
    "encoding/gob"
    "encoding/hex"
    "fmt"
)

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

// SetID sets ID of a transaction
func (tx *Transaction) SetID() error {
    var encoded bytes.Buffer
    enc := gob.NewEncoder(&encoded)
    
    err := enc.Encode(tx)
    if err != nil {
        return fmt.Errorf("failed to encode transaction: %v", err)
    }
    
    hash := sha256.Sum256(encoded.Bytes())
    tx.ID = hash[:]
    
    return nil
}

// IsCoinbase checks whether the transaction is coinbase
func (tx *Transaction) IsCoinbase() bool {
    return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
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

// CanUnlockOutputWith checks if the input can unlock an output
func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
    return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith checks if the output can be unlocked with the given data
func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
    return out.ScriptPubKey == unlockingData
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, findSpendableOutputs func(string, int) (int, map[string][]int, error)) (*Transaction, error) {
    var inputs []TxInput
    var outputs []TxOutput

    // Find spendable outputs
    acc, validOutputs, err := findSpendableOutputs(from, amount)
    if err != nil {
        return nil, fmt.Errorf("failed to find spendable outputs: %v", err)
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
                ScriptSig: from,
            }
            inputs = append(inputs, input)
        }
    }

    // Create the outputs
    outputs = append(outputs, TxOutput{
        Value:        amount,
        ScriptPubKey: to,
    })

    // If there is change, send it back to the sender
    if acc > amount {
        outputs = append(outputs, TxOutput{
            Value:        acc - amount,
            ScriptPubKey: from,
        })
    }

    tx := &Transaction{
        ID:   []byte{},
        Vin:  inputs,
        Vout: outputs,
    }

    if err := tx.SetID(); err != nil {
        return nil, fmt.Errorf("failed to set transaction ID: %v", err)
    }

    return tx, nil
}

// NewCoinbaseTx creates a new coinbase transaction
func NewCoinbaseTx(to, data string) *Transaction {
    if data == "" {
        data = fmt.Sprintf("Reward to '%s'", to)
    }

    txin := TxInput{
        Txid:      []byte{},
        Vout:      -1,
        ScriptSig: data,
    }

    txout := TxOutput{
        Value:        50, // Mining reward
        ScriptPubKey: to,
    }

    tx := &Transaction{
        ID:   []byte{},
        Vin:  []TxInput{txin},
        Vout: []TxOutput{txout},
    }

    err := tx.SetID()
    if err != nil {
        // Log error but don't return it as this is a constructor
        fmt.Printf("Error setting transaction ID: %v\n", err)
    }

    return tx
}

