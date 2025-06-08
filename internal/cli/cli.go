package cli

import (
    "flag"
    "fmt"
    "os"
    "strconv"

    "github.com/OmSingh2003/decentralized-ledger/internal/blockchain"
    "github.com/OmSingh2003/decentralized-ledger/internal/crypto/pow"
    "github.com/OmSingh2003/decentralized-ledger/internal/transaction"
    "github.com/OmSingh2003/decentralized-ledger/internal/wallet"
)

// CLI responsible for processing command line arguments
type CLI struct {
    bc *blockchain.Blockchain
}

// NewCLI creates a new CLI instance
func NewCLI(bc *blockchain.Blockchain) *CLI {
    return &CLI{bc}
}

func (cli *CLI) printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  createwallet - Creates a new wallet")
    fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
    fmt.Println("  listaddresses - Lists all addresses from the wallet file")
    fmt.Println("  printchain - Print all the blocks of the blockchain")
    fmt.Println("  reindexutxo - Rebuilds the UTXO set")
    fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// validateArgs validates command line arguments
func (cli *CLI) validateArgs() {
    if len(os.Args) < 2 {
        cli.printUsage()
        os.Exit(1)
    }
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() error {
    cli.validateArgs()

    createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
    getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
    listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
    reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
    sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

    getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
    sendFrom := sendCmd.String("from", "", "Source wallet address")
    sendTo := sendCmd.String("to", "", "Destination wallet address")
    sendAmount := sendCmd.Int("amount", 0, "Amount to send")

    switch os.Args[1] {
    case "createwallet":
        err := createWalletCmd.Parse(os.Args[2:])
        if err != nil {
            return err
        }
    case "getbalance":
        err := getBalanceCmd.Parse(os.Args[2:])
        if err != nil {
            return err
        }
    case "listaddresses":
        err := listAddressesCmd.Parse(os.Args[2:])
        if err != nil {
            return err
        }
    case "printchain":
        err := printChainCmd.Parse(os.Args[2:])
        if err != nil {
            return err
        }
    case "reindexutxo":
        err := reindexUTXOCmd.Parse(os.Args[2:])
        if err != nil {
            return err
        }
    case "send":
        err := sendCmd.Parse(os.Args[2:])
        if err != nil {
            return err
        }
    default:
        cli.printUsage()
        return fmt.Errorf("invalid command")
    }

    if createWalletCmd.Parsed() {
        return cli.createWallet()
    }

    if getBalanceCmd.Parsed() {
        if *getBalanceAddress == "" {
            getBalanceCmd.Usage()
            return fmt.Errorf("address is required")
        }
        return cli.getBalance(*getBalanceAddress)
    }

    if listAddressesCmd.Parsed() {
        return cli.listAddresses()
    }

    if printChainCmd.Parsed() {
        return cli.printChain()
    }

    if reindexUTXOCmd.Parsed() {
        return cli.reindexUTXO()
    }

    if sendCmd.Parsed() {
        if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
            sendCmd.Usage()
            return fmt.Errorf("from, to and amount are required")
        }
        return cli.send(*sendFrom, *sendTo, *sendAmount)
    }

    return nil
}

func (cli *CLI) createWallet() error {
    w := wallet.NewWallet()
    address := w.GetAddress()
    fmt.Printf("Your new address: %s\n", address)
    return nil
}

func (cli *CLI) getBalance(address string) error {
    w := wallet.LoadWallet(address)
    if w == nil {
        return fmt.Errorf("wallet not found for address: %s", address)
    }

    pubKeyHash := wallet.HashPubKey(w.PublicKey)

    UTXOSet := blockchain.UTXOSet{Blockchain: cli.bc}
    UTXOs := UTXOSet.FindUTXO(pubKeyHash)

    balance := 0
    for _, out := range UTXOs {
        balance += out.Value
    }

    fmt.Printf("Balance of '%s': %d\n", address, balance)
    return nil
}

func (cli *CLI) listAddresses() error {
    addresses := wallet.ListAddresses()
    for _, address := range addresses {
        fmt.Println(address)
    }
    return nil
}

func (cli *CLI) printChain() error {
    bci := cli.bc.Iterator()

    for {
        block, err := bci.Next()
        if err != nil {
            return fmt.Errorf("error getting next block: %v", err)
        }
        if block == nil {
            break
        }

        fmt.Printf("============ Block %x ============\n", block.Hash)
        fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
        powCheck := pow.NewProofOfWork(block)
        fmt.Printf("PoW: %s\n\n", strconv.FormatBool(powCheck.Validate()))

        for _, tx := range block.Transactions {
            fmt.Println(tx)
        }
        fmt.Printf("\n\n")

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }
    return nil
}

func (cli *CLI) reindexUTXO() error {
    UTXOSet := blockchain.UTXOSet{Blockchain: cli.bc}
    err := UTXOSet.Reindex()
    if err != nil {
        return fmt.Errorf("failed to reindex UTXO: %v", err)
    }
    
    count := len(UTXOSet.FindUTXO(nil))
    fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
    return nil
}

func (cli *CLI) send(from, to string, amount int) error {
    fromWallet := wallet.LoadWallet(from)
    if fromWallet == nil {
        return fmt.Errorf("wallet not found for address: %s", from)
    }

    toWallet := wallet.LoadWallet(to)
    if toWallet == nil {
        return fmt.Errorf("wallet not found for address: %s", to)
    }

    UTXOSet := blockchain.UTXOSet{Blockchain: cli.bc}

    tx, err := transaction.NewUTXOTransaction(fromWallet, toWallet.PublicKey, amount, UTXOSet.FindSpendableOutputs)
    if err != nil {
        return fmt.Errorf("failed to create transaction: %v", err)
    }

    cbTx := transaction.NewCoinbaseTx(fromWallet.PublicKey, "")
    txs := []*transaction.Transaction{cbTx, tx}

    newBlock, err := cli.bc.MineBlock(txs)
    if err != nil {
        return fmt.Errorf("failed to mine new block: %v", err)
    }

    err = UTXOSet.Update(newBlock)
    if err != nil {
        return fmt.Errorf("failed to update UTXO set: %v", err)
    }

    fmt.Println("Success!")
    return nil
}
