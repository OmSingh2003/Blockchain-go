package Cli

import (
	"flag"
	"fmt"
	//"log"
	"os"
	"strconv"

	"github.com/OmSingh2003/blockchain-go/blockchain"
	"github.com/OmSingh2003/blockchain-go/ProofOfWork"
	//"github.com/OmSingh2003/blockchain-go/types"
)

// CLI represents a command line interface for the blockchain
type CLI struct {
	bc *blockchain.Blockchain
}

// NewCLI creates a new CLI instance
func NewCLI(bc *blockchain.Blockchain) *CLI {
	return &CLI{bc: bc}
}

// printUsage shows the command usage
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

// validateArgs validates command line arguments
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run processes command line arguments and executes commands
func (cli *CLI) Run() error {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

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
	default:
		cli.printUsage()
		return fmt.Errorf("invalid command: %s", os.Args[1])
	}

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

	return nil
}

// addBlock adds a new block to the blockchain
func (cli *CLI) addBlock(data string) error {
	err := cli.bc.AddBlock(data)
	if err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}
	fmt.Println("Success!")
	return nil
}

// printChain prints all blocks in the blockchain
func (cli *CLI) printChain() error {
	bci, err := cli.bc.Iterator()
	if err != nil {
		return fmt.Errorf("failed to create blockchain iterator: %v", err)
	}

	for {
		block, err := bci.Next()
		if err != nil {
			return fmt.Errorf("failed to get next block: %v", err)
		}

		if block == nil {
			break
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		
		pow := ProofOfWork.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	
	return nil
}
