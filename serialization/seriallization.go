package serialization

import (
    "bytes"
    "encoding/gob"
    "github.com/OmSingh2003/blockchain-go/types"
)

func SerializeBlock(b *types.Block) []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    err := encoder.Encode(b)
    if err != nil {
        return nil
    }
    return result.Bytes()
}

func DeserializeBlock(d []byte) *types.Block {
    var block types.Block
    decoder := gob.NewDecoder(bytes.NewReader(d))
    err := decoder.Decode(&block)
    if err != nil {
        return nil
    }
    return &block
}
