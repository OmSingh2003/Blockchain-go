package main 

import (
	"fmt"
	"strconv"
	
	"github.com/OmSingh2003/blockchain-go/blockchain"
	"github.com/OmSingh2003/blockchain-go/ProofOfWork"
)

func main() {
	bc := blockchain.NewBlockchain()

	bc.AddBlock("Send 1 BTC to Ryuga")
	bc.AddBlock("Send 2 more BTC to Aztec")

	for _, block := range bc.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		
		pow := ProofOfWork.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
