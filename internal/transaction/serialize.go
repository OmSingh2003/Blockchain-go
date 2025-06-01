package transaction

import (
    "bytes"
    "encoding/gob"
    "log"
)

// SerializeOutputs serializes TxOutput array
func SerializeOutputs(outs []TxOutput) []byte {
    var buff bytes.Buffer

    enc := gob.NewEncoder(&buff)
    err := enc.Encode(outs)
    if err != nil {
        log.Panic(err)
    }

    return buff.Bytes()
}

// DeserializeOutputs deserializes TxOutput array
func DeserializeOutputs(data []byte) []TxOutput {
    var outputs []TxOutput

    dec := gob.NewDecoder(bytes.NewReader(data))
    err := dec.Decode(&outputs)
    if err != nil {
        log.Panic(err)
    }

    return outputs
}
