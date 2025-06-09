
# Variables
APP_NAME=decentralized-ledger
BINARY_NAME=blockchain
CMD_DIR=./cmd/main

# All phony targets
.PHONY: help build run test clean format lint vet check deps install create-wallet list-addresses print-chain get-balance send reindex-utxo clean-db

# Default target
.DEFAULT_GOAL := help

# Help
help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Run
run: build ## Build and run the application
	@./$(BINARY_NAME)

# Test
test: ## Run tests
	@go test -v ./...

# Clean
clean: ## Clean build artifacts
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Code quality
format: ## Format code
	@go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then golangci-lint run; else echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; fi

vet: ## Run go vet
	@go vet ./...

check: format vet test ## Run all checks

# Dependencies
deps: ## Install development dependencies
	@go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install: build ## Install binary to GOPATH/bin
	@go install $(CMD_DIR)

# Blockchain commands
create-wallet: build ## Create a new wallet
	@./$(BINARY_NAME) createwallet

list-addresses: build ## List all addresses
	@./$(BINARY_NAME) listaddresses

print-chain: build ## Print blockchain
	@./$(BINARY_NAME) printchain

get-balance: build ## Get balance (make get-balance ADDRESS=addr)
	@./$(BINARY_NAME) getbalance -address $(ADDRESS)

send: build ## Send coins (make send FROM=addr TO=addr AMOUNT=10)
	@./$(BINARY_NAME) send -from $(FROM) -to $(TO) -amount $(AMOUNT)

reindex-utxo: build ## Reindex UTXO set
	@./$(BINARY_NAME) reindexutxo

clean-db: ## Clean blockchain database
	@rm -f blockchain.db
	@rm -rf wallets/
