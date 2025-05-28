package serialization
import (
	"encoding/gob"
	
	"github.com/OmSingh2003/blockchain-go/types"
)
func (b Block.block) Searlize() []byte{
	var result bytes.Buffer 
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)

	return result.Bytes

}
func DeserializeBlock (d []byte) *Block {
	var block  Block.block

	decoder := gob.NewDecoder(bytes.NewReader(d))

	err := decoder.Decode(&Block.block)
	
	return &block
}  
