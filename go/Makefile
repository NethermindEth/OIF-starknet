.PHONY: help build run test clean deps lint

# Default target
help:
	@echo "Available targets:"
	@echo "  deps     - Install dependencies"
	@echo "  build    - Build the solver binary"
	@echo "  build-all - Build all tools (solver + network tools)"
	@echo "  run      - Run the solver"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  rebuild  - Clean and rebuild everything"
	@echo "  lint     - Run linter"
	@echo ""
	@echo "Network Setup:"
	@echo "  setup-forks              - Setup all forks (EVM and Starknet in parallel)"
	@echo "  setup-evm-forks          - Setup EVM forks (verify + setup contracts)"
	@echo "  setup-starknet-fork      - Setup Starknet fork (declare + deploy + setup)"
	@echo ""
	@echo "Network Management:"
	@echo "  start-networks           - Start all testnet forks with color-coded logging"
	@echo "  kill-networks            - Stop all running networks"
	@echo ""
	@echo "EVM Network Setup:"
	@echo "  setup-evm-contracts - Deploy ERC20 tokens to all forked EVM networks"
	@echo "  verify-evm-hyperlane   - Verify pre-deployed Hyperlane7683 contracts on EVM networks"
	@echo ""
	@echo "Starknet Network Setup:"
	@echo "  declare-sn-hyperlane7683 - Declare Hyperlane7683 contract on Starknet"
	@echo "  declare-sn-mock-erc20    - Declare MockERC20 contract on Starknet"
	@echo "  deploy-sn-hyperlane7683  - Deploy Hyperlane7683 contract to Starknet"
	@echo "  deploy-sn-mock-erc20     - Deploy MockERC20 tokens to Starknet"
	@echo "  setup-starknet-contracts - Fund users and set allowances on Starknet"
	@echo ""
	@echo "Order Management:"
	@echo "  open-random-evm-order      - Open random order on EVM (to another EVM chain)"
	@echo "  open-random-evm-order-to-sn - Open random order from EVM to Starknet"
	@echo "  open-random-sn-order       - Open random order from Starknet to EVM"

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build the solver
build:
	go build -o bin/solver ./cmd/solver

# Clean all built binaries
clean:
	@echo "🧹 Cleaning all binaries..."
	@rm -rf bin/*
	@echo "✅ Clean complete"

# Build all tools
build-all: build-verify-hyperlane build-setup-evm-contracts build-open-order build-open-starknet-order build-deploy-hyperlane7683 build-declare-hyperlane7683 build-declare-mock-erc20 build-deploy-mock-erc20 build-setup-starknet-contracts

# Rebuild everything (clean + build)
rebuild: clean build-all

# Run the solver
run: build
	./bin/solver

# Start all testnet forks with color-coded logging
start-networks:
	./start-networks.sh

# Kill all running networks
kill-networks:
	@echo "🛑 Stopping all networks..."
	@for port in 8545 8546 8547 8548; do \
		if [ -f "/tmp/anvil_$$port.pid" ]; then \
			pid=$$(cat "/tmp/anvil_$$port.pid"); \
			kill $$pid 2>/dev/null || true; \
			rm -f "/tmp/anvil_$$port.pid"; \
		fi; \
	done
	@pkill -f "anvil" 2>/dev/null || true
	@echo "✅ All networks stopped"

# Verify pre-deployed Hyperlane7683 contracts on forked networks
verify-evm-hyperlane: build-verify-hyperlane
	./bin/verify-hyperlane7683

# Deploy ERC20 tokens to all forked EVM networks
setup-evm-contracts: build-setup-evm-contracts
	./bin/setup-evm-contracts

# Deploy Hyperlane7683 contract to Starknet
deploy-sn-hyperlane7683: build-deploy-hyperlane7683
	./bin/deploy-sn-hyperlane7683

# Declare Hyperlane7683 contract on Starknet (get class hash)
declare-sn-hyperlane7683: build-declare-hyperlane7683
	./bin/declare-sn-hyperlane7683

# Declare MockERC20 contract on Starknet
declare-sn-mock-erc20: build-declare-mock-erc20
	./bin/declare-sn-mock-erc20

# Deploy MockERC20 tokens to Starknet
deploy-sn-mock-erc20: build-deploy-mock-erc20
	./bin/deploy-sn-mock-erc20

# Setup Starknet contracts (fund users and set allowances)
setup-starknet-contracts: build-setup-starknet-contracts
	./bin/setup-starknet-contracts

# Open a basic hardcoded test order on EVM
open-basic-evm-order: build-open-order
	./bin/open-order basic

# Open a random order on EVM (to another EVM chain)
open-random-evm-order: build-open-order
	./bin/open-order random

# Open a random order from EVM to Starknet
open-random-evm-order-to-sn: build-open-order
	./bin/open-order random-to-sn

# Open a random order from Starknet to EVM
open-random-sn-order: build-open-starknet-order
	./bin/open-starknet-order random

# Build Hyperlane7683 verification tool
build-verify-hyperlane:
	go build -o bin/verify-hyperlane7683 ./cmd/setup-forks/evm/verify-hyperlane7683-is-deployed

# Build token deployment tool
build-setup-evm-contracts:
	go build -o bin/setup-evm-contracts ./cmd/setup-forks/evm/setup-contracts

# Build order opening tool
build-open-order:
	go build -o bin/open-order ./cmd/open-order/evm

# Build Starknet order opening tool
build-open-starknet-order:
	go build -o bin/open-starknet-order ./cmd/open-order/starknet

# Build Hyperlane7683 deployment tool
build-deploy-hyperlane7683:
	go build -o bin/deploy-sn-hyperlane7683 ./cmd/setup-forks/starknet/deploy/hyperlane7683

# Build Hyperlane7683 declaration tool
build-declare-hyperlane7683:
	go build -o bin/declare-sn-hyperlane7683 ./cmd/setup-forks/starknet/declare/hyperlane7683

# Build MockERC20 declaration tool
build-declare-mock-erc20:
	go build -o bin/declare-sn-mock-erc20 ./cmd/setup-forks/starknet/declare/mock-erc20

# Build MockERC20 deployment tool
build-deploy-mock-erc20:
	go build -o bin/deploy-sn-mock-erc20 ./cmd/setup-forks/starknet/deploy/mock-erc20

# Build Starknet contract setup tool
build-setup-starknet-contracts:
	go build -o bin/setup-starknet-contracts ./cmd/setup-forks/starknet/setup-contracts

# Setup EVM forks (verify hyperlane + setup contracts)
setup-evm-forks: verify-evm-hyperlane setup-evm-contracts
	@echo "✅ EVM forks setup complete!"

# Setup Starknet fork (declare + deploy + setup contracts)
setup-starknet-fork: declare-sn-hyperlane7683 declare-sn-mock-erc20 deploy-sn-hyperlane7683 deploy-sn-mock-erc20 setup-starknet-contracts
	@echo "✅ Starknet fork setup complete!"

# Setup all forks (EVM and Starknet in parallel)
setup-forks: setup-evm-forks setup-starknet-fork
	@echo "🎉 All forks setup complete!"

# Run tests
test:
	go test ./...

# Run linter
lint:
	golangci-lint run

# Install development tools
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

# Run with hot reload (requires air)
dev:
	air

# Generate mocks (if using mockery)
mocks:
	mockery --all --keeptree --output=./internal/mocks
