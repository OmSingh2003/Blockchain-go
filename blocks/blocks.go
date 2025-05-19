package blocks
import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)
// A block in block chain usually have these information 
type Block struct {
	Timestamp        int64 // Records when block was created/mined
	Data             []byte // Payload of the block : Actual Information 
	PrevBlockHash    []byte // Stores the Hash of previous Block int the chain 
	Hash             []byte // Stores the Hash of current block in the chain
}
// made this block to compute unique cryptographic hash for the Block 'b' and store it in b.haash
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp,10))
	// Converts int64 of timestamp to []bytes [first to string (decimal that is base of 10)]
	// it is converted because these crptographic algorithms work on slices of bytes not on int64 or strings.
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp},[]byte{})  // PREPARING INPUT FOR HASING ALGORITHM 
	// bytes.Join is used to concatanate the slice of bytes with separator  in between 
	// List containing the three byte sequences : []byte{} is used as sepatator which is an empty slice
	// no seprator in between [one more thing order is very important]
	hash := sha256.Sum256(headers) // IMPLEMENTING sha256 as hashing algorithm 
	// SHA256 takes a single continous stream of data thats why i concatanated it 
	// return an array of 32bytes 
	b.Hash = hash[:] // Assigning value of hash to the b block
	// [:] create a slice of bytes when applied to an array 
}
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(), // Set current time as Timestamp
		Data:          []byte(data),      // Convert input string data to []byte
		PrevBlockHash: prevBlockHash,     // Assign the hash of the previous block
		Hash:          []byte{},          // Initialize Hash as an empty byte slice 
	}
	block.SetHash() //Calculate the hash for the block using the SetHash Function that I made just above this func 
	return block // return the block 
}

