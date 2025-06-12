# Decentralized Ledger

[![Tests](https://github.com/OmSingh2003/decentralized-ledger/actions/workflows/test.yml/badge.svg)](https://github.com/OmSingh2003/decentralized-ledger/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24.2-00ADD8?style=flat&logo=go)](https://golang.org/)

A complete decentralized ledger system written in Go, featuring multiple consensus algorithms (PoW and PoS), wallet management, and transaction processing with UTXO (Unspent Transaction Output) model.

## Features

### Core Blockchain Features
- **Multiple Consensus Algorithms**:
  - **Proof of Work (PoW)**: Secure mining algorithm with dynamic difficulty adjustment
  - **Proof of Stake (PoS)**: Energy-efficient consensus with validator signatures (In Development)
- **UTXO Model**: Bitcoin-like transaction model for efficient balance tracking
- **Dynamic Difficulty Adjustment**: Automatic mining difficulty adjustment based on block time
- **Merkle Trees**: Efficient transaction verification and integrity
- **Persistent Storage**: BoltDB for reliable blockchain data storage

### Wallet & Transaction Management
- **Wallet Management**: Create and manage multiple wallets with cryptographic key pairs
- **Transaction Processing**: Send and receive coins between addresses
- **Digital Signatures**: ECDSA-based transaction signing and verification
- **Address Generation**: Base58 encoding with checksum validation

### Developer Experience
- **CLI Interface**: Comprehensive command-line interface for blockchain interaction
- **Modular Architecture**: Clean separation of concerns with pluggable consensus
- **Comprehensive Testing**: Unit tests for critical components
- **CI/CD Pipeline**: Automated testing with GitHub Actions
- **Makefile**: Convenient build and development commands

## Architecture

The project follows a clean modular architecture:

```
├── cmd/blockchain/          # Main application entry point
├── internal/
│   ├── blockchain/          # Core blockchain logic and UTXO management
│   ├── block/              # Block structure and operations
│   ├── cli/                # Command-line interface
│   ├── crypto/
│   │   ├── pow/            # Proof of Work implementation
│   │   └── merkletree/     # Merkle tree for transaction verification
│   ├── transaction/        # Transaction creation and validation
│   └── wallet/             # Wallet and cryptographic operations
├── pkg/serialization/      # Data serialization utilities
└── wallets/               # Wallet storage directory
```

## Prerequisites

- Go 1.24.2 or higher
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/OmSingh2003/decentralized-ledger.git
cd decentralized-ledger
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o decentralized-ledger cmd/blockchain/main.go
```

## Quick Start

### 1. Create a Wallet

First, create a wallet to receive mining rewards:

```bash
./decentralized-ledger createwallet
```

This will output your new wallet address. Save this address as you'll need it for the next step.

### 2. Initialize the Blockchain

Create the genesis block and initialize the blockchain:

```bash
./decentralized-ledger init -address YOUR_WALLET_ADDRESS
```

Replace `YOUR_WALLET_ADDRESS` with the address from step 1.

### 3. Check Your Balance

Check the balance of your wallet (should show mining reward from genesis block):

```bash
./decentralized-ledger getbalance -address YOUR_WALLET_ADDRESS
```

### 4. Create Another Wallet

Create a second wallet to test transactions:

```bash
./decentralized-ledger createwallet
```

### 5. Send Coins

Send coins from your first wallet to the second:

```bash
./decentralized-ledger send -from SENDER_ADDRESS -to RECEIVER_ADDRESS -amount 10
```

## Available Commands

### Wallet Management

- `createwallet` - Creates a new wallet and returns its address
- `listaddresses` - Lists all wallet addresses
- `getbalance -address ADDRESS` - Get balance of a specific address

### Blockchain Operations

- `init -address ADDRESS` - Initialize blockchain with genesis block
- `printchain` - Print all blocks in the blockchain
- `send -from FROM -to TO -amount AMOUNT` - Send coins between addresses
- `reindexutxo` - Rebuild the UTXO (Unspent Transaction Output) set

### Examples

```bash
# Create a new wallet
./decentralized-ledger createwallet

# Initialize blockchain
./decentralized-ledger init -address 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa

# Check balance
./decentralized-ledger getbalance -address 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa

# Send 10 coins
./decentralized-ledger send -from 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa -to 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2 -amount 10

# View the entire blockchain
./decentralized-ledger printchain

# List all wallet addresses
./decentralized-ledger listaddresses

# Rebuild UTXO set (if needed)
./decentralized-ledger reindexutxo
```

## Technical Details

### Proof of Work

The decentralized ledger uses a SHA-256 based proof-of-work algorithm. Miners must find a nonce that, when combined with block data, produces a hash with a specific number of leading zeros.

### UTXO Model

Follows Bitcoin's UTXO (Unspent Transaction Output) model:
- Each transaction consumes previous UTXOs as inputs
- Creates new UTXOs as outputs
- Enables efficient balance calculation and double-spend prevention

### Cryptography

- **Digital Signatures**: ECDSA (Elliptic Curve Digital Signature Algorithm)
- **Hashing**: SHA-256 for block hashes and proof-of-work
- **Address Generation**: Base58 encoding with checksum

### Storage

- **Database**: BoltDB for persistent storage
- **Files**: 
  - `blockchain.db` - Main blockchain database
  - `wallets/` - Directory containing wallet files

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

The codebase is organized into several packages:

- **cmd**: Application entry points
- **internal**: Private application code
  - **blockchain**: Core blockchain functionality
  - **block**: Block data structures
  - **crypto**: Cryptographic operations (PoW, Merkle trees)
  - **transaction**: Transaction handling
  - **wallet**: Wallet management
  - **cli**: Command-line interface
- **pkg**: Public library code

### Dependencies

- `go.etcd.io/bbolt` - Embedded key-value database
- `golang.org/x/crypto` - Extended cryptography library

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is open source and available under the [MIT License](LICENSE).

## Acknowledgments

- Inspired by Bitcoin's blockchain design
- Built following Go best practices and clean architecture principles

## Troubleshooting

### Common Issues

1. **"Wallet not found" error**: Make sure you've created a wallet using `createwallet` before trying to use an address.

2. **"Insufficient funds" error**: Check your balance with `getbalance` before sending transactions.

3. **Database errors**: If you encounter database issues, ensure you have write permissions in the project directory.

4. **Build errors**: Make sure you're using Go 1.24.2 or higher and run `go mod tidy` to install dependencies.

### Reset Blockchain

To start fresh, delete the database file:

```bash
rm blockchain.db
```

Then reinitialize with the `init` command.
