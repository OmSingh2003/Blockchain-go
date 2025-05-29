package transactions
import (

)
//in outputs coin are stored
type transactions struct {
	ID      []byte
	Vin     []TXInput 
	Vout    []TXOutput
}

type TXOutput struct {
	Value        int // coins 
	ScriptPubKey string // Locking them with a help of puzzle // avoiding scripting right now 
}
type TXInput struct {
	TXid       []byte // stores the id of transactions 
	Vout       int  // index of the output 
	ScriptSig  string // provides data to be used in ScriptPubKey
}
