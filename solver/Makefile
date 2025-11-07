
.PHONY: help build run run-local run-live test test-unit test-rpc-local test-rpc-live test-integration-local test-integration-live test-solver-local test-solver-live test-all test-coverage test-coverage-html test-coverage-check test-coverage-all clean deps dev-deps lint kill-all fund-accounts fund-accounts-local fund-accounts-live register-starknet-on-evm register-starknet-on-evm-local register-starknet-on-evm-live start-networks check-networks-local kill-networks open-random-evm-order-local open-random-evm-order-live open-random-evm-sn-order-local open-random-evm-sn-order-live open-random-sn-order-local open-random-sn-order-live

# Default target
help:
	@echo "🚀 OIF-Starknet - Open Intent Framework with Starknet Support"
	@echo ""
	@echo "📋 Testing Commands:"
	@echo "  test-unit        - Run unit tests (no RPC required)"
	@echo "  test-rpc-local   - Run RPC tests with local devnet (requires start-networks)"
	@echo "  test-rpc-live    - Run RPC tests with live networks"
	@echo "  test-integration-local - Run basic integration tests with local devnet (network setup and order opening)"
	@echo "  test-integration-live - Run basic integration tests with live testnets (network setup and order opening)"
	@echo "  test-solver-local - Run solver integration tests with local devnet (multi-order processing)"
	@echo "  test-solver-live - Run solver integration tests with live testnets (multi-order processing)"
	@echo "  test-coverage    - Show coverage for maintainable code"
	@echo ""
	@echo "🔧 Development Commands:"
	@echo "  deps             - Install dependencies"
	@echo "  dev-deps         - Install development tools (linter, debugger)"
	@echo "  build            - Build the solver binary"
	@echo "  run              - Run the solver (uses current IS_DEVNET setting)"
	@echo "  run-local        - Run the solver with local devnet (IS_DEVNET=true)"
	@echo "  run-live         - Run the solver with live networks (IS_DEVNET=false)"
	@echo "  clean            - Clean build artifacts"
	@echo "  clean-solver     - Clean solver state (keep binaries)"
	@echo "  lint             - Run linter"
	@echo ""
	@echo "🌐 Network Management:"
	@echo "  start-networks   - Start all testnet forks (runs continuously)"
	@echo "  kill-networks    - Stop all running networks"
	@echo "  kill-all         - Stop all processes (networks + solvers) and clean state"
	@echo "  fund-accounts    - Fund Alice & Solver with tokens (uses current IS_DEVNET setting)"
	@echo "  fund-accounts-local - Fund Alice & Solver on local devnet (IS_DEVNET=true)"
	@echo "  fund-accounts-live - Fund Alice & Solver on live networks (IS_DEVNET=false)"
	@echo "  register-starknet-on-evm - Register Starknet domain on EVM contracts (uses current IS_DEVNET setting)"
	@echo "  register-starknet-on-evm-local - Register Starknet domain on EVM contracts (local devnet)"
	@echo "  register-starknet-on-evm-live - Register Starknet domain on EVM contracts (live networks)"
	@echo ""
	@echo "🎯 Order Management:"
	@echo "  open-random-evm-order-local - Open random EVM order with local devnet (IS_DEVNET=true)"
	@echo "  open-random-evm-order-live - Open random EVM order with live networks (IS_DEVNET=false)"
	@echo "  open-random-evm-sn-order-local - Open random EVM-to-Starknet order with local devnet (IS_DEVNET=true)"
	@echo "  open-random-evm-sn-order-live - Open random EVM-to-Starknet order with live networks (IS_DEVNET=false)"
	@echo "  open-random-sn-order-local - Open random Starknet order with local devnet (IS_DEVNET=true)"
	@echo "  open-random-sn-order-live - Open random Starknet order with live networks (IS_DEVNET=false)"
	@echo ""
	@echo "📊 Coverage Commands:"
	@echo "  test-coverage    - Show coverage for maintainable code"
	@echo "  test-coverage-html - Generate HTML coverage report"
	@echo "  test-coverage-check - Check coverage thresholds"
	@echo "  test-coverage-all - Show coverage for all packages"
	@echo ""
	@echo "🔍 Advanced Testing:"
	@echo "  test-pkg PKG=./path     - Run tests for specific package"
	@echo "  test-pattern PATTERN=TestName - Run tests matching pattern"
	@echo ""
	@echo "📦 Build Commands:"
	@echo "  build-all        - Build all common tools"
	@echo "  build-extra      - Build all tools (deployment, verification, etc.)"
	@echo "  rebuild          - Clean and rebuild everything"


### Available Targets ###

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install development dependencies (linter, debugger, etc.)
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

# Run the solver (uses current IS_DEVNET setting)
run: build
	./bin/solver solver

# Run the solver with local devnet (sets IS_DEVNET=true)
run-local: build
	@echo "🤖 Starting solver with local devnet (IS_DEVNET=true)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
	IS_DEVNET=true ./bin/solver solver

# Run the solver with live networks (sets IS_DEVNET=false)
run-live: build
	@echo "🤖 Starting solver with live networks (IS_DEVNET=false)..."
	@echo "Make sure you have live network access and proper environment variables set"
	IS_DEVNET=false ./bin/solver solver

# Build the solver
build:
	go build -o bin/solver ./cmd/main.go

# Build all necessary tools for common use and setup
build-all: build build-fund-accounts build-register-evm-routers

# Build all tools including setup, deployment, verification, etc.
build-extra: build-all build-verify-hyperlane build-deploy-hyperlane7683 build-declare-hyperlane7683 build-declare-mock-erc20 build-deploy-mock-erc20 build-deploy-forge-mock-erc20 build-setup-starknet-contracts build-register-sn-routers

# Rebuild everything (clean + build)
rebuild: clean build-all

# Clean all built binaries
clean:
	@echo "🧹 Cleaning all binaries..."
	@rm -rf bin/*
	@echo "🧹 Cleaning solver state..."
	@rm -f state/solver_state/solver-state.json solver-state.json
	@echo "✅ Clean complete (solver will start from .env start blocks)"

# Clean all built binaries
clean-solver:
	@echo "🧹 Cleaning solver state..."
	@rm -f state/solver_state/solver-state.json solver-state.json
	@echo "✅ Clean complete (solver will start from .env start blocks)"

# Run linter
lint:
	golangci-lint run

# Run linter with fix suggestions
lint-fix:
	golangci-lint run --fix

# Run linter on specific package
lint-pkg:
	golangci-lint run ./solvercore/...

# testing related targets at the bottom

### Network Management: ###

# Check if local networks are running (used by test targets)
check-networks-local:
	@echo "🔍 Checking if local networks are running..."
	@failed=0; \
	if ! curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' http://localhost:8545 >/dev/null 2>&1; then \
		echo "❌ Ethereum (port 8545) is not accessible"; \
		failed=1; \
	fi; \
	if ! curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' http://localhost:8546 >/dev/null 2>&1; then \
		echo "❌ Optimism (port 8546) is not accessible"; \
		failed=1; \
	fi; \
	if ! curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' http://localhost:8547 >/dev/null 2>&1; then \
		echo "❌ Arbitrum (port 8547) is not accessible"; \
		failed=1; \
	fi; \
	if ! curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' http://localhost:8548 >/dev/null 2>&1; then \
		echo "❌ Base (port 8548) is not accessible"; \
		failed=1; \
	fi; \
	if ! curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"starknet_chainId","params":[],"id":1}' http://localhost:5050 >/dev/null 2>&1; then \
		echo "❌ Starknet (port 5050) is not accessible"; \
		failed=1; \
	fi; \
	if [ $$failed -eq 1 ]; then \
		echo ""; \
		echo "❌ One or more networks are not running!"; \
		echo ""; \
		echo "💡 To start all networks, run in a separate terminal:"; \
		echo "   make start-networks"; \
		echo ""; \
		echo "💡 Or check individual network status:"; \
		echo "   - Ethereum:   curl http://localhost:8545"; \
		echo "   - Optimism:  curl http://localhost:8546"; \
		echo "   - Arbitrum:  curl http://localhost:8547"; \
		echo "   - Base:      curl http://localhost:8548"; \
		echo "   - Starknet:  curl http://localhost:5050"; \
		echo ""; \
		exit 1; \
	fi; \
	echo "✅ All networks are running and accessible"

# Start all testnet forks with color-coded logging
start-networks:
	./cmd/tools/start-networks.sh

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

# Kill all solver processes and networks
kill-all:
	@echo "🧹 Cleaning up all processes..."
	@echo "🛑 Stopping solver processes..."
	@pkill -f "solver solver" 2>/dev/null || true
	@pkill -f "./bin/solver solver" 2>/dev/null || true
	@echo "🛑 Stopping all networks..."
	@for port in 8545 8546 8547 8548; do \
		if [ -f "/tmp/anvil_$$port.pid" ]; then \
			pid=$$(cat "/tmp/anvil_$$port.pid"); \
			kill $$pid 2>/dev/null || true; \
			rm -f "/tmp/anvil_$$port.pid"; \
		fi; \
	done
	@pkill -f "anvil" 2>/dev/null || true
	@pkill -f "katana" 2>/dev/null || true
	@echo "🧹 Cleaning solver state..."
	@rm -f solvercore/config/solver-state.json 2>/dev/null || true
	@rm -f solvercore/solvers/hyperlane7683/solver-state.json 2>/dev/null || true
	@echo "✅ All processes cleaned up"

# Note: setup-networks removed because start-networks runs continuously
# Use separate terminals:
# Terminal 1: make start-networks
# Terminal 2: make fund-accounts register-starknet-on-evm

### Individual Setup Commands (use as needed) ###

# Fund Alice and Solver accounts with MockERC20 tokens (uses current IS_DEVNET setting)
fund-accounts: build-fund-accounts
	@if [ -z "$(NETWORK)" ]; then \
		echo "Funding Alice & Solver on all networks..."; \
		./bin/fund-accounts all $(AMOUNT); \
	else \
		echo "Funding Alice & Solver on $(NETWORK)..."; \
		./bin/fund-accounts $(NETWORK) $(AMOUNT); \
	fi

# Fund Alice and Solver accounts with MockERC20 tokens (local devnet)
fund-accounts-local: build-fund-accounts
	@echo "Funding Alice & Solver on local devnet (IS_DEVNET=true)..."
	@IS_DEVNET=true ./bin/fund-accounts all $(AMOUNT)

# Fund Alice and Solver accounts with MockERC20 tokens (live networks)
fund-accounts-live: build-fund-accounts
	@echo "Funding Alice & Solver on live networks (IS_DEVNET=false)..."
	@IS_DEVNET=false ./bin/fund-accounts all $(AMOUNT)

# Essential setup for testing (register Starknet domain on EVM contracts - uses current IS_DEVNET setting)
register-starknet-on-evm: build-register-evm-routers
	./bin/register-evm-routers
	@echo "✅ Starknet domain registered on all EVM contracts!"

# Register Starknet domain on EVM contracts (local devnet)
register-starknet-on-evm-local: build-register-evm-routers
	@echo "Registering Starknet domain on EVM contracts (local devnet)..."
	IS_DEVNET=true ./bin/register-evm-routers
	@echo "✅ Starknet domain registered on all EVM contracts!"

# Register Starknet domain on EVM contracts (live networks)
register-starknet-on-evm-live: build-register-evm-routers
	@echo "Registering Starknet domain on EVM contracts (live networks)..."
	IS_DEVNET=false ./bin/register-evm-routers
	@echo "✅ Starknet domain registered on all EVM contracts!"

# Not required for testing if forking post deployment (register EVM domains on Starknet contract)
register-evm-on-starknet: build-register-sn-routers
	./bin/register-sn-routers
	@echo "✅ EVM domains registered on Starknet contract!"

# Build account funding tool
build-fund-accounts:
	go build -o bin/fund-accounts ./cmd/tools/fund-accounts

# Build helper tools
build-register-evm-routers:
	go build -o bin/register-evm-routers ./cmd/tools/additional-helpers/register-evm-routers

build-register-sn-routers:
	go build -o bin/register-sn-routers ./cmd/tools/additional-helpers/register-sn-routers

### More Setup Commands (use as needed) ###

# Verify pre-deployed Hyperlane7683 contracts on forked networks
verify-evm-hyperlane: build-verify-hyperlane
	./bin/verify-hyperlane7683

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

# Deploy ERC20 tokens to all forked EVM networks

# Setup Starknet contracts (fund users and set allowances)
setup-starknet-contracts: build-setup-starknet-contracts
	./bin/setup-starknet-contracts

# Build Hyperlane7683 verification tool
build-verify-hyperlane:
	go build -o bin/verify-hyperlane7683 ./cmd/tools/additional-helpers/verify-hyperlane7683

# Deploy MockERC20 with Forge (guarantees verification works)
deploy-forge-mock-erc20: build-deploy-forge-mock-erc20
	@if [ -z "$(NETWORK)" ]; then \
		echo "Deploying MockERC20 with Forge to all EVM networks..."; \
		./bin/deploy-forge-mock-erc20; \
	else \
		echo "Deploying MockERC20 with Forge to $(NETWORK)..."; \
		./bin/deploy-forge-mock-erc20 $(NETWORK); \
	fi

# Build Hyperlane7683 deployment tool
build-deploy-hyperlane7683:
	go build -o bin/deploy-sn-hyperlane7683 ./cmd/tools/additional-helpers/deploy-sn-hyperlane7683

# Build Hyperlane7683 declaration tool
build-declare-hyperlane7683:
	go build -o bin/declare-sn-hyperlane7683 ./cmd/tools/additional-helpers/declare-sn-hyperlane7683

# Build MockERC20 declaration tool
build-declare-mock-erc20:
	go build -o bin/declare-sn-mock-erc20 ./cmd/tools/additional-helpers/declare-sn-mock-erc20

# Build MockERC20 deployment tool
build-deploy-mock-erc20:
	go build -o bin/deploy-sn-mock-erc20 ./cmd/tools/additional-helpers/deploy-sn-mock-erc20

# Build MockERC20 Forge deployment tool
build-deploy-forge-mock-erc20:
	go build -o bin/deploy-forge-mock-erc20 ./cmd/tools/deploy-forge-mock-erc20

# Build token deployment tool
# Note: build-setup-evm-contracts target removed - no corresponding source file found

# Build Starknet contract setup tool
build-setup-starknet-contracts:
	go build -o bin/setup-starknet-contracts ./cmd/tools/additional-helpers/setup-starknet-contracts

### Order Management: ###

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

# Open random EVM order with local devnet (sets IS_DEVNET=true)
open-random-evm-order-local: build
	@echo "🎯 Opening random EVM order with local devnet (IS_DEVNET=true)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
	IS_DEVNET=true ./bin/solver tools open-order evm

# Open random EVM order with live networks (sets IS_DEVNET=false)
open-random-evm-order-live: build
	@echo "🎯 Opening random EVM order with live networks (IS_DEVNET=false)..."
	@echo "Make sure you have live network access and proper environment variables set"
	IS_DEVNET=false ./bin/solver tools open-order evm

# Open random EVM-to-Starknet order with local devnet (sets IS_DEVNET=true)
open-random-evm-sn-order-local: build
	@echo "🎯 Opening random EVM-to-Starknet order with local devnet (IS_DEVNET=true)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
	IS_DEVNET=true ./bin/solver tools open-order evm random-to-sn

# Open random EVM-to-Starknet order with live networks (sets IS_DEVNET=false)
open-random-evm-sn-order-live: build
	@echo "🎯 Opening random EVM-to-Starknet order with live networks (IS_DEVNET=false)..."
	@echo "Make sure you have live network access and proper environment variables set"
	IS_DEVNET=false ./bin/solver tools open-order evm random-to-sn

# Open random Starknet order with local devnet (sets IS_DEVNET=true)
open-random-sn-order-local: build
	@echo "🎯 Opening random Starknet order with local devnet (IS_DEVNET=true)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
	IS_DEVNET=true ./bin/solver tools open-order starknet

# Open random Starknet order with live networks (sets IS_DEVNET=false)
open-random-sn-order-live: build
	@echo "🎯 Opening random Starknet order with live networks (IS_DEVNET=false)..."
	@echo "Make sure you have live network access and proper environment variables set"
	IS_DEVNET=false ./bin/solver tools open-order starknet

### Testing Commands ###

# Run unit tests (no RPC required)
test-unit:
	@echo "🧪 Running unit tests (no RPC required)..."
	@SKIP_RPC_TESTS=true go test -v -short \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-run "Test.*" \
		-timeout 30s

# Run RPC tests with local devnet (requires start-networks)
test-rpc-local: check-networks-local
	@echo "🌐 Running RPC tests with local devnet..."
	@set -a && . ./.env && set +a && IS_DEVNET=true go test -v -count=1 \
		./pkg/starknetutil \
		./pkg/ethutil \
		-run "TestMaintainableWithNetworks|TestRPCCoverageWithNetworks" \
		-timeout 60s

# Run RPC tests with live networks (requires live network access)
test-rpc-live:
	@echo "🌐 Running RPC tests with live networks..."
	@echo "Make sure you have live network access and proper environment variables set"
	@set -a && . ./.env && set +a && IS_DEVNET=false go test -v -count=1 \
		./pkg/starknetutil \
		./pkg/ethutil \
		-run "TestMaintainableWithNetworks|TestRPCCoverageWithNetworks" \
		-timeout 60s

# Run basic integration tests with local devnet (order opening only)
test-integration-local: check-networks-local
	@echo "🔄 Running basic integration tests with local devnet (order opening only)..."
	@echo "Funding accounts and setting up contracts..."
	@make fund-accounts-local
	@make register-starknet-on-evm-local
	@echo "Starting integration tests..."
	@set -a && . ./.env && set +a && IS_DEVNET=true EXECUTE_ORDER_COMMANDS=true go test -v -count=1 -run "TestOrderCreationCommandsIntegration" -timeout 300s -p 1

# Run basic integration tests with live testnets (order opening only)
test-integration-live:
	@echo "🔄 Running basic integration tests with live testnets (order opening only)..."
	@echo "Make sure you have live network access and proper environment variables set"
	@set -a && . ./.env && set +a && IS_DEVNET=false EXECUTE_ORDER_COMMANDS=true go test -v -count=1 -run "TestOrderCreationCommandsIntegration" -timeout 300s -p 1

# Run solver integration tests with local devnet (order opening + solving)
test-solver-local: check-networks-local
	@echo "🤖 Running solver integration tests with local devnet (multi-order processing)..."
	@echo "Funding accounts and setting up contracts..."
	@make fund-accounts-local
	@make register-starknet-on-evm-local
	@echo "Starting multi-order solver integration test..."
	@set -a && . ./.env && set +a && IS_DEVNET=true EXECUTE_SOLVER_TESTS=true go test -v -count=1 -run "TestSolverIntegration/CompleteOrderLifecycle_MultiOrder" -timeout 400s -p 1

# Run solver integration tests with live testnets (order opening + solving)
test-solver-live:
	@echo "🤖 Running solver integration tests with live testnets (multi-order processing)..."
	@echo "Make sure you have live network access and proper environment variables set"
	@set -a && . ./.env && set +a && IS_DEVNET=false EXECUTE_SOLVER_TESTS=true go test -v -count=1 -run "TestSolverIntegration/CompleteOrderLifecycle_MultiOrder" -timeout 400s -p 1

# Run all tests (unit + RPC + integration)
# Check networks first so errors appear early, but also show summary at end if needed
test-all: check-networks-local test-unit test-rpc-local test-integration-local
	@echo "✅ All tests completed!"

# Show coverage for maintainable code
test-coverage:
	@echo "📊 Running coverage analysis for maintainable code..."
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-coverprofile=coverage.out
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1

# Generate HTML coverage report
test-coverage-html:
	@echo "📊 Generating HTML coverage report..."
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check coverage thresholds (simplified)
test-coverage-check:
	@echo "📊 Checking coverage thresholds..."
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-coverprofile=coverage.out
	@echo ""
	@echo "Coverage by package:"
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 | grep -E "coverage:" | while read line; do \
		pkg=$$(echo "$$line" | awk '{print $$1}'); \
		coverage=$$(echo "$$line" | grep -o 'coverage: [0-9.]*%' | cut -d' ' -f2); \
		echo "  - $$pkg: $$coverage"; \
	done
	@echo ""
	@echo "Overall coverage:"
	@go tool cover -func=coverage.out | tail -1

# Run tests with coverage for all packages (including cmd tools)
test-coverage-all:
	@echo "📊 Running coverage analysis for all packages..."
	@go test -cover ./... -coverprofile=coverage.all.out || true
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.all.out | tail -1

# Run tests for specific package
test-pkg:
	@echo "Usage: make test-pkg PKG=./pkg/starknetutil"
	@if [ -z "$(PKG)" ]; then echo "Please specify PKG=./path/to/package"; exit 1; fi
	go test -v $(PKG)

# Run tests matching a pattern
test-pattern:
	@echo "Usage: make test-pattern PATTERN=TestBalance"
	@if [ -z "$(PATTERN)" ]; then echo "Please specify PATTERN=TestName"; exit 1; fi
	go test -v ./... -run $(PATTERN)


# # Run with hot reload (requires air)
# dev:
# 	air
#
# # Generate mocks (if using mockery)
# mocks:
# 	mockery --all --keeptree --output=./solvercore/mocks

