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
	fmt.Println("  addblock -data DATA -miner ADDRESS - add a block to the blockchain and reward the miner")
	fmt.Println("  printchain - print all the blocks of the blockchain")
	fmt.Println("  getbalance -address ADDRESS - get balance of ADDRESS")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - send AMOUNT of coins from FROM address to TO")
}

// validateArgs validates the command line arguments
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// send sends coins from one address to another
func (cli *CLI) send(from, to string, amount int) error {
    // Create a new transaction using a closure to pass the FindSpendableOutputs method
    tx, err := transactions.NewUTXOTransaction(
        from,
        to,
        amount,
        func(address string, amount int) (int, map[string][]int, error) {
            return cli.Bc.FindSpendableOutputs(address, amount)
        },
    )
    if err != nil {
        return fmt.Errorf("failed to create transaction: %v", err)
    }

    // Add the transaction to a new block
    err = cli.Bc.AddBlock([]*transactions.Transaction{tx})
    if err != nil {
        return fmt.Errorf("failed to add block: %v", err)
    }

    fmt.Println("Success! Transaction has been added to the blockchain")
    return nil
}

// addBlock adds a block to the blockchain
func (cli *CLI) addBlock(data string, minerAddress string) error {
	// Create a coinbase transaction for the miner
	coinbaseTx := transactions.NewCoinbaseTx(minerAddress, "")

	// Create a data transaction if data is provided
	var txs []*transactions.Transaction
	txs = append(txs, coinbaseTx)

	if data != "" {
		dataTx := &transactions.Transaction{
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
		err := dataTx.SetID()
		if err != nil {
			return fmt.Errorf("failed to set transaction ID: %v", err)
		}

		txs = append(txs, dataTx)
	}

	// Add the transactions to the blockchain
	err := cli.Bc.AddBlock(txs)
	if err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}

	fmt.Printf("Block mined! Miner %s received the reward.\n", minerAddress)
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
	UTXOs, err := cli.Bc.FindUTXO(address)
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
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	// Define flags for commands
	addBlockData := addBlockCmd.String("data", "", "Block data")
	addBlockMiner := addBlockCmd.String("miner", "", "Miner address to receive the reward")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	
	// Send command flags
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

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
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			return fmt.Errorf("failed to parse send command: %v", err)
		}
	default:
		cli.printUsage()
		return fmt.Errorf("invalid command: %s", os.Args[1])
	}

	// Execute the appropriate command
	if addBlockCmd.Parsed() {
		if *addBlockMiner == "" {
			addBlockCmd.Usage()
			return fmt.Errorf("miner address is required")
		}
		return cli.addBlock(*addBlockData, *addBlockMiner)
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

	if sendCmd.Parsed() {
		if *sendFrom == "" {
			sendCmd.Usage()
			return fmt.Errorf("from address is required")
		}
		if *sendTo == "" {
			sendCmd.Usage()
			return fmt.Errorf("to address is required")
		}
		if *sendAmount <= 0 {
			sendCmd.Usage()
			return fmt.Errorf("amount must be greater than 0")
		}
		return cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	return nil
}
