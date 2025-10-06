.PHONY: help build run test clean deps lint

# Default target
help:
	@echo "🚀 OIF-Starknet - Open Intent Framework with Starknet Support"
	@echo ""
	@echo "📋 Testing Commands:"
	@echo "  test-unit        - Run unit tests (no RPC required)"
	@echo "  test-solver-integration - Run full solver tests (executes orders + solver)"
	@echo "  test-all         - Run all tests (unit + RPC + integration)"
	@echo "  test-rpc-local   - Run RPC tests with local devnet (requires start-networks)"
	@echo "  test-rpc-live    - Run RPC tests with live networks"
	@echo "  test-integration-local - Run basic integration tests with local devnet (network setup and order opening)"
	@echo "  test-integration-live - Run basic integration tests with live testnets (network setup and order opening)"
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
	@echo "📊 Coverage Commands:"
	@echo "  test-coverage    - Show coverage for maintainable code"
	@echo "  test-coverage-with-integration - Show coverage including integration tests"
	@echo "  coverage-help    - Show all coverage commands"
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
	@if [ -z "$(NETWORK)" ]; then \
		IS_DEVNET=true ./bin/fund-accounts all $(AMOUNT); \
	else \
		IS_DEVNET=true ./bin/fund-accounts $(NETWORK) $(AMOUNT); \
	fi

# Fund Alice and Solver accounts with MockERC20 tokens (live networks)
fund-accounts-live: build-fund-accounts
	@echo "Funding Alice & Solver on live networks (IS_DEVNET=false)..."
	@if [ -z "$(NETWORK)" ]; then \
		IS_DEVNET=false ./bin/fund-accounts all $(AMOUNT); \
	else \
		IS_DEVNET=false ./bin/fund-accounts $(NETWORK) $(AMOUNT); \
	fi

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

### Testing Commands ###

# Run unit tests (no RPC required)
test-unit:
	@echo "🧪 Running unit tests (no RPC required)..."
	@go test -v \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-run "Test.*" -timeout 30s

# Run RPC tests with local devnet (requires start-networks)
test-rpc-local:
	@echo "🌐 Running RPC tests with local devnet (requires start-networks)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
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
test-integration-local:
	@echo "🔄 Running basic integration tests with local devnet (order opening only)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
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

# Run full solver integration tests (requires full setup + solver execution)
test-solver-integration:
	@echo "🤖 Running full solver integration tests (requires full setup)..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
	@echo "And setup is complete with: make fund-accounts register-starknet-on-evm"
	@EXECUTE_SOLVER_TESTS=true go test -v -run "TestSolverIntegration" -timeout 400s -p 1

# Run all tests (unit + RPC + integration)
test-all: test-unit test-rpc test-integration
	@echo "✅ All tests completed!"

# Show coverage for maintainable code
test-coverage:
	@echo "📊 Running coverage analysis for maintainable code..."
	@make test-coverage-maintainable

# Show coverage including integration tests
test-coverage-with-integration:
	@echo "📊 Running coverage analysis including integration tests..."
	@echo "This includes unit tests + integration tests for more accurate coverage"
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-coverprofile=coverage.unit.out
	@echo ""
	@echo "Running integration tests with coverage..."
	@go test -cover -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" \
		-coverprofile=coverage.integration.out
	@echo ""
	@echo "Combining coverage reports..."
	@echo "mode: set" > coverage.combined.out
	@tail -n +2 coverage.unit.out >> coverage.combined.out
	@tail -n +2 coverage.integration.out >> coverage.combined.out
	@echo ""
	@echo "Combined coverage summary:"
	@go tool cover -func=coverage.combined.out | tail -1

# Show available coverage commands
coverage-help:
	@echo "📊 Available coverage commands:"
	@echo ""
	@echo "  make test-coverage                    - Show coverage for maintainable code"
	@echo "  make test-coverage-with-integration   - Show coverage including integration tests"
	@echo "  make test-coverage-unit               - Show coverage for unit tests only"
	@echo "  make test-coverage-rpc                - Show coverage for RPC tests only"
	@echo "  make test-coverage-all                - Show coverage for all packages"
	@echo "  make test-coverage-html               - Generate HTML coverage report"
	@echo ""
# Note: test-coverage-html is defined below in the Additional Coverage Commands section

# Run tests with coverage for all packages
test-coverage-all:
	@echo "Running coverage analysis for all packages..."
	@go test -cover ./... -coverprofile=coverage.out
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "Coverage by package:"
	@go test -cover ./... | grep -E "coverage:" | while read line; do \
		pkg=$$(echo "$$line" | awk '{print $$1}'); \
		coverage=$$(echo "$$line" | grep -o 'coverage: [0-9.]*%' | cut -d' ' -f2); \
		echo "  - $$pkg: $$coverage"; \
	done

# Run tests with coverage excluding auto-generated and untestable packages
test-coverage-core:
	@echo "Running coverage analysis for core packages (excluding auto-generated and cmd tools)..."
	@go test -cover ./pkg/... ./solvercore/... -coverprofile=coverage.core.out
	@echo ""
	@echo "Core coverage summary:"
	@go tool cover -func=coverage.core.out | tail -1
	@echo ""
	@echo "Core coverage by package:"
	@go test -cover ./pkg/... ./solvercore/... | grep -E "coverage:" | while read line; do \
		pkg=$$(echo "$$line" | awk '{print $$1}'); \
		coverage=$$(echo "$$line" | grep -o 'coverage: [0-9.]*%' | cut -d' ' -f2); \
		echo "  - $$pkg: $$coverage"; \
	done

# Run tests with coverage for maintainable code only (excludes contracts, cmd tools, and generated code)
test-coverage-maintainable:
	@echo "Running coverage analysis for maintainable code only..."
	@echo "Excluding: contracts (auto-generated), cmd tools, and other untestable packages"
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-coverprofile=coverage.maintainable.out
	@echo ""
	@echo "Maintainable code coverage summary:"
	@go tool cover -func=coverage.maintainable.out | tail -1
	@echo ""
	@echo "Maintainable code coverage by package:"
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

# Run tests with network coverage (requires running networks)
test-coverage-networks:
	@echo "Running coverage analysis with network tests..."
	@echo "Note: This requires running networks and may take longer"
	@echo "Set SKIP_RPC_TESTS=true to skip network-dependent tests"
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-coverprofile=coverage.networks.out
	@echo ""
	@echo "Network coverage summary:"
	@go tool cover -func=coverage.networks.out | tail -1
	@echo ""
	@echo "Network coverage by package:"
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

# Check coverage thresholds
test-coverage-check:
	@echo "Checking coverage thresholds..."
	@echo "Running comprehensive coverage analysis..."
	@go test -cover ./... -coverprofile=coverage.out
	@echo ""
	@echo "Coverage analysis by package:"
	@go test -cover ./... | grep -E "coverage:" | while read line; do \
		pkg=$$(echo "$$line" | awk '{print $$1}'); \
		coverage=$$(echo "$$line" | grep -o 'coverage: [0-9.]*%' | cut -d' ' -f2 | cut -d'%' -f1); \
		if [ "$$(echo "$$coverage < 50" | bc -l)" = "1" ]; then \
			echo "❌ $$pkg: $$coverage% (Target: 50%+)"; \
		elif [ "$$(echo "$$coverage < 70" | bc -l)" = "1" ]; then \
			echo "⚠️  $$pkg: $$coverage% (Target: 70%+)"; \
		elif [ "$$(echo "$$coverage < 80" | bc -l)" = "1" ]; then \
			echo "✅ $$pkg: $$coverage% (Target: 80%+)"; \
		else \
			echo "🎉 $$pkg: $$coverage% (Target: 80%+)"; \
		fi; \
	done
	@echo ""
	@echo "Overall coverage:"
	@go tool cover -func=coverage.out | tail -1

# Note: test-unit and test-integration are defined above in the Testing Commands section

# Run integration tests with local forks
test-integration-local:
	@echo "Running integration tests with local forks..."
	@echo "Make sure local forks are running with: make start-networks"
	IS_DEVNET=true go test -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v

# Run integration tests with live testnets
test-integration-live:
	@echo "Running integration tests with live testnets..."
	@echo "Make sure you have valid API keys in your .env file"
	IS_DEVNET=false go test -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v

# Run integration tests with coverage (includes order creation commands)
test-integration-coverage:
	@echo "Running integration tests with coverage..."
	@echo "This will cover order creation code paths that are missing from unit tests"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set IS_DEVNET=true for local forks or IS_DEVNET=false for live testnets"
	go test -cover -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v -p 1

# Run integration tests with coverage and order command execution
test-integration-coverage-full:
	@echo "Running full integration tests with coverage (includes order creation commands)..."
	@echo "This will execute actual order creation commands to cover missing code paths"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set IS_DEVNET=true for local forks or IS_DEVNET=false for live testnets"
	EXECUTE_ORDER_COMMANDS=true go test -cover -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios|TestOrderCreationCommandsIntegration" -v -p 1

# Run comprehensive order creation integration tests with balance verification
test-order-creation-integration:
	@echo "Running comprehensive order creation integration tests..."
	@echo "This will test order creation with full balance verification across all networks"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set IS_DEVNET=true for local forks or IS_DEVNET=false for live testnets"
	EXECUTE_ORDER_COMMANDS=true go test -run "TestOrderCreationCommandsIntegration" -v -p 1

# Run comprehensive order creation integration tests with coverage
test-order-creation-integration-coverage:
	@echo "Running comprehensive order creation integration tests with coverage..."
	@echo "This will test order creation with full balance verification and coverage analysis"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set IS_DEVNET=true for local forks or IS_DEVNET=false for live testnets"
	EXECUTE_ORDER_COMMANDS=true go test -cover -run "TestOrderCreationCommandsIntegration" -v -p 1

# Note: test-solver-integration is defined above in the Testing Commands section

# Run complete order lifecycle integration tests with coverage
test-solver-integration-coverage:
	@echo "Running complete order lifecycle integration tests with coverage..."
	@echo "This will test the full order lifecycle with coverage analysis"
	@echo "Tests all 3 order types: EVM→EVM, EVM→Starknet, Starknet→EVM"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set IS_DEVNET=true for local forks or IS_DEVNET=false for live testnets"
	EXECUTE_SOLVER_TESTS=true go test -cover -run "TestSolverIntegration" -v -p 1

### Additional Coverage Commands ###

# Show coverage for unit tests only
test-coverage-unit:
	@echo "📊 Running coverage analysis for unit tests..."
	@go test -cover \
		./pkg/envutil \
		./pkg/starknetutil \
		./pkg/ethutil \
		./solvercore/base \
		./solvercore/config \
		./solvercore/logutil \
		./solvercore/types \
		./solvercore/solvers/hyperlane7683 \
		-run "Test.*" \
		-coverprofile=coverage.unit.out
	@echo ""
	@echo "Unit test coverage summary:"
	@go tool cover -func=coverage.unit.out | tail -1

# Show coverage for RPC tests only
test-coverage-rpc:
	@echo "📊 Running coverage analysis for RPC tests..."
	@echo "Make sure networks are running in another terminal with: make start-networks"
	@go test -cover \
		./pkg/starknetutil \
		./pkg/ethutil \
		-run "TestMaintainableWithNetworks|TestRPCCoverageWithNetworks" \
		-coverprofile=coverage.rpc.out
	@echo ""
	@echo "RPC test coverage summary:"
	@go tool cover -func=coverage.rpc.out | tail -1

# Generate HTML coverage report
test-coverage-html:
	@echo "📊 Generating HTML coverage report..."
	@go test -cover ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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


