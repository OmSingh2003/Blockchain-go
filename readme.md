# Simple Blockchain in Go

This project implements a basic blockchain from scratch using Go. It includes a command-line interface (CLI) for interacting with the blockchain, supporting wallet management, transactions, block mining, and blockchain inspection.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Building the Project](#building-the-project)
- [Usage](#usage)
  - [Creating a Wallet](#creating-a-wallet)
  - [Listing Addresses](#listing-addresses)
  - [Validating an Address](#validating-an-address)
  - [Getting Balance](#getting-balance)
  - [Mining a Block](#mining-a-block)
  - [Sending Coins](#sending-coins)
  - [Printing the Blockchain](#printing-the-blockchain)

## Features

- **Proof-of-Work (PoW)**: Implements a simple PoW algorithm to ensure mining difficulty and blockchain security.
- **Transactions**: Follows the UTXO (Unspent Transaction Output) model to support coin transfers.
- **Wallets**: Enables ECDSA-based wallet creation, address generation, and transaction signing.
- **Persistence**: Uses `bbolt` (BoltDB) for local blockchain storage and data persistence.
- **CLI Interface**: Provides a command-line tool for performing blockchain operations.

## Project Structure

The project is organized into modular packages:

- `blockchain/Cli`: Parses and handles CLI commands.
- `blockchain/ProofOfWork`: Contains the PoW consensus logic.
- `blockchain/blockchain`: Core logic for adding blocks, handling UTXOs, and database interaction.
- `blockchain/transactions`: Handles transaction structure, signing, and verification.
- `blockchain/types`: Defines the `Block` structure and related utilities.
- `blockchain/wallet`: Manages key generation, address encoding/validation, and wallet storage.
- `blockchain/merkleTree`: Implements a Merkle Tree for transaction verification (optional/advanced).
- `blockchain/serialization`: Utility functions for data serialization/deserialization.
- `blockchain/versions`: (Optional) Intended for versioning/network sync functionality.

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Building the Project

1. Clone the repository:
    ```bash
    git clone <repository_url>
    cd blockchain-go  # or your project directory
    ```

2. Build the CLI executable:
    ```bash
    go build -o blockchain-cli main.go
    ```

This will generate an executable named `blockchain-cli` in your current directory.

## Usage

All operations are performed using the `blockchain-cli` executable.

Make sure you are in the directory containing `blockchain-cli`.

### Creating a Wallet

Generates a new wallet and prints its address. Wallets are saved to `wallet.dat`.

```bash
./blockchain-cli createwallet

Listing Addresses

Displays all addresses stored in the wallet.dat file.

./blockchain-cli listaddresses

Validating an Address

Checks whether a given address is valid.

./blockchain-cli validateaddress -address <ADDRESS>

Example:

./blockchain-cli validateaddress -address 1EUTWhURsSm2pB4SWEcedyuCegCGt6GTx1

Getting Balance

Retrieves the balance of a given address by summing all its UTXOs.

./blockchain-cli getbalance -address <ADDRESS>

Example:

./blockchain-cli getbalance -address 193BfVq7YWesFsds9fNE3RzoqqwV2krqou

Mining a Block

Mines a new block, including a coinbase transaction to the miner’s address. You can also include optional data in the block.

./blockchain-cli addblock -miner <MINER_ADDRESS> [-data <DATA>]

Example:

./blockchain-cli addblock -miner 193BfVq7YWesFsds9fNE3RzoqqwV2krqou -data "First transaction!"

Sending Coins

Creates a new transaction to send coins from one address to another and mines it into a block.

./blockchain-cli send -from <FROM_ADDRESS> -to <TO_ADDRESS> -amount <AMOUNT>

Example:

./blockchain-cli send -from 193BfVq7YWesFsds9fNE3RzoqqwV2krqou -to 1EUTWhURsSm2pB4SWEcedyuCegCGt6GTx1 -amount 20

Printing the Blockchain

Prints the entire blockchain, displaying block hashes, previous hashes, timestamps, nonces, and transactions.

./blockchain-cli printchain


⸻

License

This project is open source and available under the MIT License.

