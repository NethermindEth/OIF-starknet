.PHONY: help build run test clean deps lint

# Default target
help:
	@echo "Available targets:"
	@echo "  deps     - Install dependencies"
	@echo "  dev-deps - Install development tools (linter, debugger)"
	@echo "  build    - Build the solver binary"
	@echo "  build-all - Build all tools (solver + network tools)"
	@echo "  run      - Run the solver"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  rebuild  - Clean and rebuild everything"
	@echo "  lint     - Run linter"
	@echo "  lint-fix - Run linter with auto-fix"
	@echo ""
	@echo "Network Setup:"
	@echo "  setup-forks              - Setup all forks (EVM and Starknet)"
	@echo "  setup-evm-forks          - Setup EVM forks"
	@echo "  setup-starknet-fork      - Setup Starknet fork"
	@echo ""
	@echo "Network Management:"
	@echo "  start-networks           - Start all testnet forks with logging"
	@echo "  kill-networks            - Stop all running networks"
	@echo ""
	@echo "EVM Network Setup:"
	@echo "  setup-evm-contracts      - Deploy ERC20 tokens to all forked EVM networks and fund accounts"
	@echo "  verify-evm-hyperlane     - Verify pre-deployed Hyperlane7683 contracts on each EVM network"
	@echo "  register-evm-routers     - Register starknet domain each EVM hyperlane contract (impersonated owner)"
	@echo ""
	@echo "Starknet Network Setup:"
	@echo "  declare-sn-hyperlane7683 - Declare Hyperlane7683 contract on Starknet"
	@echo "  declare-sn-mock-erc20    - Declare MockERC20 contract on Starknet"
	@echo "  deploy-sn-hyperlane7683  - Deploy Hyperlane7683 contract to Starknet"
	@echo "  deploy-sn-mock-erc20     - Deploy MockERC20 tokens to Starknet"
	@echo "  setup-starknet-contracts - Fund Starknet users"
	@echo "  register-sn-routers      - Register EVM domains on Starknet"
	@echo ""
	@echo "Order Management:"
	@echo "  open-random-evm-order    - Open random order on an EVM chain (to fill on another EVM chain)"
	@echo "  open-random-evm-sn-order - Open random order from an EVM chain to Starknet"
	@echo "  open-random-sn-order     - Open random order from Starknet to a random EVM chain"
	@echo ""
	@echo "Default Order Commands (for debugging - identical data except nonces):"
	@echo "  default-evm-order                   - Open default EVM→EVM order"
	@echo "  default-evm-sn-order                - Open default EVM→Starknet order"
	@echo "  default-sn-order                    - Open default Starknet→EVM order"

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install development dependencies (linter, debugger, etc.)
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

# Build the solver
build:
	go build -o bin/solver ./cmd/main.go

# Clean all built binaries and solver state
clean:
	@echo "🧹 Cleaning all binaries..."
	@rm -rf bin/*
	@echo "🧹 Cleaning solver state..."
	@rm -f state/solver_state/solver-state.json solver-state.json
	@echo "✅ Clean complete (solver will start from .env start blocks)"

# Build all tools
build-all: build build-verify-hyperlane build-setup-evm-contracts build-deploy-hyperlane7683 build-declare-hyperlane7683 build-declare-mock-erc20 build-deploy-mock-erc20 build-setup-starknet-contracts build-register-evm-routers build-register-sn-routers

# Rebuild everything (clean + build)
rebuild: clean build-all

# Run the solver
run: build
	./bin/solver solver

# Start all testnet forks with color-coded logging
start-networks:
	./cmd/setup-forks/start-networks.sh

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

# Build helper tools
build-register-evm-routers:
	go build -o bin/register-evm-routers ./cmd/setup-forks/evm/setup-contracts/register-routers

build-register-sn-routers:
	go build -o bin/register-sn-routers ./cmd/setup-forks/starknet/setup-contracts/register_routers

# Default order commands for debugging (identical data except nonces)
default-evm-order: build
	./bin/solver tools open-order evm

default-evm-sn-order: build
	./bin/solver tools open-order evm

default-sn-order: build
	./bin/solver tools open-order starknet

open-random-evm-order: build
	./bin/solver tools open-order evm

open-random-evm-sn-order: build
	./bin/solver tools open-order evm random-to-sn

open-random-sn-order: build
	./bin/solver tools open-order starknet

# Build Hyperlane7683 verification tool
build-verify-hyperlane:
	go build -o bin/verify-hyperlane7683 ./cmd/setup-forks/evm/verify-hyperlane7683-is-deployed

# Build token deployment tool
build-setup-evm-contracts:
	go build -o bin/setup-evm-contracts ./cmd/setup-forks/evm/setup-contracts

# Order opening is now handled by the single binary
# build-open-order and build-open-starknet-order removed - use ./bin/solver tools open-order <chain>

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
setup-evm-forks: verify-evm-hyperlane setup-evm-contracts build-register-evm-routers
	./bin/register-evm-routers
	@echo "✅ EVM forks setup complete!"

# Setup Starknet fork (declare + deploy + setup contracts)
setup-starknet-fork: declare-sn-mock-erc20 deploy-sn-mock-erc20 setup-starknet-contracts
	@echo "✅ Starknet fork setup complete!"
# # Setup Starknet fork (declare + deploy + setup contracts)
# setup-starknet-fork: declare-sn-hyperlane7683 declare-sn-mock-erc20 deploy-sn-hyperlane7683 deploy-sn-mock-erc20 setup-starknet-contracts build-register-sn-routers
# 	./bin/register-sn-routers
# 	@echo "✅ Starknet fork setup complete!"

# Setup all forks (Starknet first so EVM can reference SN address), then EVM
setup-forks: setup-starknet-fork setup-evm-forks
	@echo "🎉 All forks setup complete!"

# Run tests
test:
	go test ./...

# Run linter
lint:
	golangci-lint run

# Run linter with fix suggestions
lint-fix:
	golangci-lint run --fix

# Run linter on specific package
lint-pkg:
	golangci-lint run ./solvercore/...



# Run with hot reload (requires air)
dev:
	air

# Generate mocks (if using mockery)
mocks:
	mockery --all --keeptree --output=./solvercore/mocks
