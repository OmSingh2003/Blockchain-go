package Cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/OmSingh2003/blockchain-go/blockchain"
	"github.com/OmSingh2003/blockchain-go/ProofOfWork"
	"github.com/OmSingh2003/blockchain-go/transactions"
)

// CLI responsible for processing command line arguments
type CLI struct {
	Bc *blockchain.Blockchain
}

// NewCLI creates a new CLI instance
func NewCLI(bc *blockchain.Blockchain) *CLI {
	return &CLI{Bc: bc}
}

// printUsage prints the usage of the CLI
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
	fmt.Println("  getbalance -address ADDRESS - get balance of ADDRESS")
}

// validateArgs validates the command line arguments
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// addBlock adds a block to the blockchain
func (cli *CLI) addBlock(data string) error {
	// Create a simple transaction with the data
	tx := &transactions.Transaction{
		ID: []byte{},
		Vin: []transactions.TxInput{
			{
				Txid:      []byte{},
				Vout:      -1,
				ScriptSig: data,
			},
		},
		Vout: []transactions.TxOutput{
			{
				Value:        0,
				ScriptPubKey: "data",
			},
		},
	}

	// Set the transaction ID
	err := tx.SetID()
	if err != nil {
		return fmt.Errorf("failed to set transaction ID: %v", err)
	}

	// Add the transaction to the blockchain
	err = cli.Bc.AddBlock([]*transactions.Transaction{tx})
	if err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}

	fmt.Println("Block added successfully!")
	return nil
}

// printChain prints all the blocks in the blockchain
func (cli *CLI) printChain() error {
	// Create an iterator for the blockchain
	bci, err := cli.Bc.Iterator()
	if err != nil {
		return fmt.Errorf("failed to create blockchain iterator: %v", err)
	}

	for {
		// Get the next block from the iterator
		block, err := bci.Next()
		if err != nil {
			return fmt.Errorf("failed to get next block: %v", err)
		}

		if block == nil {
			break
		}

		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("Transactions: %d\n", len(block.Transactions))
		
		for i, tx := range block.Transactions {
			fmt.Printf("  Transaction %d: %x\n", i, tx.ID)
			fmt.Printf("    Inputs: %d\n", len(tx.Vin))
			fmt.Printf("    Outputs: %d\n", len(tx.Vout))
		}
		
		pow := ProofOfWork.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}

	return nil
}

// getBalance gets the balance of the specified address
func (cli *CLI) getBalance(address string) error {
	bc, err := blockchain.NewBlockchain()
	if err != nil {
		return fmt.Errorf("failed to create blockchain: %v", err)
	}
	defer bc.CloseDB()

	UTXOs, err := bc.FindUTXO(address)
	if err != nil {
		return fmt.Errorf("failed to find UTXO: %v", err)
	}

	balance := 0
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
	return nil
}

// Run processes command line arguments and executes the appropriate command
func (cli *CLI) Run() error {
	cli.validateArgs()

	// Create new flagsets for each command
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	// Define flags for commands
	addBlockData := addBlockCmd.String("data", "", "Block data")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	// Parse the appropriate command
	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			return fmt.Errorf("failed to parse addblock command: %v", err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			return fmt.Errorf("failed to parse printchain command: %v", err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			return fmt.Errorf("failed to parse getbalance command: %v", err)
		}
	default:
		cli.printUsage()
		return fmt.Errorf("invalid command: %s", os.Args[1])
	}

	// Execute the appropriate command
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			return fmt.Errorf("data flag is required")
		}

		return cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		return cli.printChain()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			return fmt.Errorf("address flag is required")
		}
		
		return cli.getBalance(*getBalanceAddress)
	}

	return nil
}
