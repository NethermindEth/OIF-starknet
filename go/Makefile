.PHONY: help build run test clean deps lint

# Default target
help:
	@echo "Available targets:"
	@echo "  deps           - Install dependencies"
	@echo "  dev-deps       - Install development tools (linter, debugger)"
	@echo "  run            - Run the solver"
	@echo "  build          - Build the solver binary"
	@echo "  build-all      - Build all common tools (for testing and setup)"
	@echo "  build-full     - Build all tools (for deployment, verification, etc.)"
	@echo "  rebuild        - Clean and rebuild everything"
	@echo "  clean          - Clean build artifacts"
	@echo "  clean-state    - Clean solver state (last indexed blocsk)"
	@echo "  lint           - Run linter"
	@echo "  lint-fix       - Run linter with auto-fix"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  test-unit      - Run only unit tests"
	@echo "  test-integration               - Run integration tests (requires networks)"
	@echo "  test-integration-coverage      - Run integration tests with coverage"
	@echo "  test-integration-coverage-full - Run full integration tests with coverage (includes order commands)"
	@echo "  test-order-creation-integration - Run comprehensive order creation tests with balance verification"
	@echo "  test-order-creation-integration-coverage - Run order creation tests with coverage and balance verification"
	@echo "  test-solver-integration          - Run complete order lifecycle tests (Open → Fill → Settle)"
	@echo "  test-solver-integration-coverage - Run complete order lifecycle tests with coverage"
	@echo "  test-pkg PKG=./path - Run tests for specific package"
	@echo "  test-pattern PATTERN=TestName - Run tests matching pattern"
	@echo ""
	@echo "Network Management:"
	@echo "  start-networks           - Start all testnet forks with logging"
	@echo "  kill-networks            - Stop all running networks"
	@echo ""
	@echo ""
	@echo "Individual Setup Commands (use as needed):"
	@echo "  fund-accounts            - Fund Alice & Solver with MockERC20MockERC20 tokens on all networks"
	@echo "  register-starknet-on-evm - Register Starknet domain on all EVM Hyperlane7683 contracts"
	@echo "  register-evm-on-starknet - Register all EVM domains on Starknet Hyperlane7683 contract"
	@echo ""
	@echo "More Setup Commands (use as needed):"
	@echo "  verify-evm-hyperlane     - Verify pre-deployed Hyperlane7683 contracts on each EVM network"
	@echo "  setup-evm-contracts      - Deploy ERC20 tokens to all forked EVM networks and fund accounts"
	@echo "  declare-sn-hyperlane7683 - Declare Hyperlane7683 contract on Starknet"
	@echo "  deploy-sn-hyperlane7683  - Deploy Hyperlane7683 contract to Starknet"
	@echo "  declare-sn-mock-erc20    - Declare MockERC20 contract on Starknet"
	@echo "  deploy-sn-mock-erc20     - Deploy MockERC20 tokens to Starknet"
	@echo "  deploy-forge-mock-erc20  - Deploy MockERC20 with Forge (guarantees verification works)"
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


### Available Targets ###

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install development dependencies (linter, debugger, etc.)
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

# Run the solver
run: build
	./bin/solver solver

# Build the solver
build:
	go build -o bin/solver ./cmd/main.go

# Build all necessary tools for common use and setup
build-all: build build-fund-accounts build-register-evm-routers build-register-sn-routers

# Build all tools including setup, deployment, verification, etc.
build-full: build build-verify-hyperlane build-setup-evm-contracts build-deploy-hyperlane7683 build-declare-hyperlane7683 build-declare-mock-erc20 build-deploy-mock-erc20 build-deploy-forge-mock-erc20 build-setup-starknet-contracts build-register-evm-routers build-register-sn-routers

# Rebuild everything (clean + build)
rebuild: clean build-all

# Clean all built binaries
clean:
	@echo "🧹 Cleaning all binaries..."
	@rm -rf bin/*
	@echo "✅ Clean complete"

# Clean only solver state (keep binaries)
clean-state:
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

### Individual Setup Commands (use as needed) ###

# Fund Alice and Solver accounts with MockERC20 tokens
fund-accounts: build-fund-accounts
	@if [ -z "$(NETWORK)" ]; then \
		echo "Funding Alice & Solver on all networks..."; \
		./bin/fund-accounts all $(AMOUNT); \
	else \
		echo "Funding Alice & Solver on $(NETWORK)..."; \
		./bin/fund-accounts $(NETWORK) $(AMOUNT); \
	fi

# Essential setup for testing (register Starknet domain on EVM contracts)
register-starknet-on-evm: build-register-evm-routers
	./bin/register-evm-routers
	@echo "✅ Starknet domain registered on all EVM contracts!"

# Not required for testing (register EVM domains on Starknet contract)
register-evm-on-starknet: build-register-sn-routers
	./bin/register-sn-routers
	@echo "✅ EVM domains registered on Starknet contract!"

# Build account funding tool
build-fund-accounts:
	go build -o bin/fund-accounts ./cmd/tools/fund-accounts

# Build helper tools
build-register-evm-routers:
	go build -o bin/register-evm-routers ./cmd/tools/setup-forks/evm/setup-contracts/register-routers

build-register-sn-routers:
	go build -o bin/register-sn-routers ./cmd/tools/setup-forks/starknet/setup-contracts/register_routers

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
setup-evm-contracts: build-setup-evm-contracts
	./bin/setup-evm-contracts

# Setup Starknet contracts (fund users and set allowances)
setup-starknet-contracts: build-setup-starknet-contracts
	./bin/setup-starknet-contracts

# Build Hyperlane7683 verification tool
build-verify-hyperlane:
	go build -o bin/verify-hyperlane7683 ./cmd/tools/setup-forks/evm/verify-hyperlane7683-is-deployed

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
	go build -o bin/deploy-sn-hyperlane7683 ./cmd/tools/setup-forks/starknet/deploy/hyperlane7683

# Build Hyperlane7683 declaration tool
build-declare-hyperlane7683:
	go build -o bin/declare-sn-hyperlane7683 ./cmd/tools/setup-forks/starknet/declare/hyperlane7683

# Build MockERC20 declaration tool
build-declare-mock-erc20:
	go build -o bin/declare-sn-mock-erc20 ./cmd/tools/setup-forks/starknet/declare/mock-erc20

# Build MockERC20 deployment tool
build-deploy-mock-erc20:
	go build -o bin/deploy-sn-mock-erc20 ./cmd/tools/setup-forks/starknet/deploy/mock-erc20

# Build MockERC20 Forge deployment tool
build-deploy-forge-mock-erc20:
	go build -o bin/deploy-forge-mock-erc20 ./cmd/tools/deploy-forge-mock-erc20

# Build token deployment tool
build-setup-evm-contracts:
	go build -o bin/setup-evm-contracts ./cmd/tools/setup-forks/evm/setup-contracts

# Build Starknet contract setup tool
build-setup-starknet-contracts:
	go build -o bin/setup-starknet-contracts ./cmd/tools/setup-forks/starknet/setup-contracts

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

### Testing Related Targets: ###

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Show available coverage commands
coverage-help:
	@echo "Available coverage commands:"
	@echo ""
	@echo "  make test-coverage-all        - Run coverage for ALL packages (includes cmd tools, contracts, etc.)"
	@echo "  make test-coverage-core       - Run coverage for pkg/ and solvercore/ only (excludes cmd tools)"
	@echo "  make test-coverage-maintainable - Run coverage for maintainable code only (excludes auto-generated)"
	@echo "  make test-coverage-check      - Check coverage thresholds with color-coded status"
	@echo "  make test-coverage-html       - Generate HTML coverage report for all packages"
	@echo ""
# Run tests with detailed coverage report
test-coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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

# Run only unit tests (exclude integration tests)
test-unit:
	go test -v ./...

# Run integration tests (requires network setup)
test-integration:
	@echo "Running integration tests..."
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	go test -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v

# Run integration tests with local forks
test-integration-local:
	@echo "Running integration tests with local forks..."
	@echo "Make sure local forks are running with: make start-networks"
	FORKING=true go test -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v

# Run integration tests with live testnets
test-integration-live:
	@echo "Running integration tests with live testnets..."
	@echo "Make sure you have valid API keys in your .env file"
	FORKING=false go test -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v

# Run integration tests with coverage (includes order creation commands)
test-integration-coverage:
	@echo "Running integration tests with coverage..."
	@echo "This will cover order creation code paths that are missing from unit tests"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	go test -cover -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios" -v

# Run integration tests with coverage and order command execution
test-integration-coverage-full:
	@echo "Running full integration tests with coverage (includes order creation commands)..."
	@echo "This will execute actual order creation commands to cover missing code paths"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	EXECUTE_ORDER_COMMANDS=true go test -cover -run "TestOrderLifecycleIntegration|TestCrossChainOperations|TestErrorScenarios|TestPerformanceScenarios|TestOrderCreationCommandsIntegration" -v

# Run comprehensive order creation integration tests with balance verification
test-order-creation-integration:
	@echo "Running comprehensive order creation integration tests..."
	@echo "This will test order creation with full balance verification across all networks"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	EXECUTE_ORDER_COMMANDS=true go test -run "TestOrderCreationCommandsIntegration" -v

# Run comprehensive order creation integration tests with coverage
test-order-creation-integration-coverage:
	@echo "Running comprehensive order creation integration tests with coverage..."
	@echo "This will test order creation with full balance verification and coverage analysis"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	EXECUTE_ORDER_COMMANDS=true go test -cover -run "TestOrderCreationCommandsIntegration" -v

# Run complete order lifecycle integration tests (Open → Fill → Settle)
test-solver-integration:
	@echo "Running complete order lifecycle integration tests..."
	@echo "This will test the full order lifecycle: Open → Fill → Settle"
	@echo "Tests all 3 order types: EVM→EVM, EVM→Starknet, Starknet→EVM"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	EXECUTE_SOLVER_TESTS=true go test -run "TestSolverIntegration" -v

# Run complete order lifecycle integration tests with coverage
test-solver-integration-coverage:
	@echo "Running complete order lifecycle integration tests with coverage..."
	@echo "This will test the full order lifecycle with coverage analysis"
	@echo "Tests all 3 order types: EVM→EVM, EVM→Starknet, Starknet→EVM"
	@echo "Make sure networks are running with: make start-networks"
	@echo "Set FORKING=true for local forks or FORKING=false for live testnets"
	EXECUTE_SOLVER_TESTS=true go test -cover -run "TestSolverIntegration" -v

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


