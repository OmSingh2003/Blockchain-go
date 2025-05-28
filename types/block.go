package types

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64  // Records when block was created/mined
	Data          []byte // Payload of the block : Actual Information 
	PrevBlockHash []byte // Stores the Hash of previous Block in the chain 
	Hash          []byte // Stores the Hash of current block in the chain
	Nonce         int    // Number used in proof of work
}

// Serialize serializes the block
func (b *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}
	
	return result.Bytes(), nil
}

// DeserializeBlock deserializes a block
func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, err
	}
	
	return &block, nil
}
// NewBlock creates a new Block with given data and previous block hash
func NewBlock(data string, prevBlockHash []byte) *Block {
	return &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
}

// PrepareData prepares data for hashing by concatenating block data with nonce
func (b *Block) PrepareData(nonce int, targetBits int64) []byte {
	data := bytes.Join(
		[][]byte{
			b.PrevBlockHash,
			b.Data,
			IntToHex(b.Timestamp),
			IntToHex(targetBits),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	return []byte(strconv.FormatInt(num, 10))
}

