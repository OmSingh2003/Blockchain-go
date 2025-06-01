package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    
    "github.com/OmSingh2003/blockchain-go/internal/blockchain"
    "github.com/OmSingh2003/blockchain-go/internal/cli"
    "github.com/OmSingh2003/blockchain-go/internal/wallet"
)

func main() {
    // Check for commands that don't require blockchain initialization
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "createwallet":
            // Create a new wallet
            w := wallet.NewWallet()
            address := w.GetAddress()
            fmt.Printf("Your new address: %s\n", address)
            return
            
        case "listaddresses":
            // List all wallet addresses
            addresses := wallet.ListAddresses()
            for _, address := range addresses {
                fmt.Println(address)
            }
            return
            
        case "init":
            // Initialize blockchain with genesis block
            initCmd := flag.NewFlagSet("init", flag.ExitOnError)
            initAddress := initCmd.String("address", "", "The address to use for mining the genesis block")
            
            if err := initCmd.Parse(os.Args[2:]); err != nil {
                log.Fatalf("Failed to parse init command: %v", err)
            }
            
            if *initAddress == "" {
                fmt.Println("Error: Address is required")
                fmt.Println("Usage: blockchain init -address WALLET_ADDRESS")
                return
            }
            
            // Validate wallet exists
            minerWallet := wallet.LoadWallet(*initAddress)
            if minerWallet == nil {
                fmt.Printf("Error: Wallet not found for address: %s\n", *initAddress)
                return
            }
            
            // Create blockchain with genesis block
            bc, err := createBlockchain(*initAddress)
            if err != nil {
                log.Fatalf("Failed to create blockchain: %v", err)
            }
            defer bc.CloseDB()
            
            fmt.Println("Blockchain initialized with genesis block!")
            return
        }
    }
    
    // For all other commands, initialize blockchain
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
    cli := cli.NewCLI(bc)
    if err := cli.Run(); err != nil {
        log.Fatalf("CLI error: %v", err)
    }
}

// createBlockchain creates a new blockchain with a genesis block and rewards the miner
func createBlockchain(minerAddress string) (*blockchain.Blockchain, error) {
    // Load the wallet for the miner - this will be checked again in CreateBlockchain
    // but we do it here first to provide a better error message
    minerWallet := wallet.LoadWallet(minerAddress)
    if minerWallet == nil {
        return nil, fmt.Errorf("wallet not found for address: %s", minerAddress)
    }
    
    // Create a new blockchain with the genesis block
    bc, err := blockchain.CreateBlockchain(minerWallet)
    if err != nil {
        return nil, fmt.Errorf("failed to create blockchain: %v", err)
    }
    
    return bc, nil
}
