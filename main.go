package main 

import (
	//"fmt"
	"log"
	
	"github.com/OmSingh2003/blockchain-go/blockchain"
	"github.com/OmSingh2003/blockchain-go/Cli"
)

func main() {
	// Initialize blockchain with proper error handling
	bc, err := blockchain.NewBlockchain()
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}
	
	// Ensure database is closed properly when main exits
	defer func() {
		if err := bc.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()
	
	// Initialize and run CLI
	cli := Cli.NewCLI(bc)
	if err := cli.Run(); err != nil {
		log.Fatalf("CLI error: %v", err)
	}
}
