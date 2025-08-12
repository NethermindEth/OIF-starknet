.PHONY: help build run test clean deps lint

# Default target
help:
	@echo "Available targets:"
	@echo "  deps     - Install dependencies"
	@echo "  build    - Build the solver binary"
	@echo "  run      - Run the solver"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  lint     - Run linter"
	@echo "  verify-hyperlane  - Verify pre-deployed Hyperlane7683 contracts"
	@echo "  deploy-tokens     - Deploy ERC20 tokens to all forked networks"
	@echo "  start-networks - Start all testnet forks with color-coded logging"
	@echo "  kill-networks  - Stop all running networks"

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build the solver
build:
	go build -o bin/solver ./cmd/solver

# Run the solver
run: build
	./bin/solver

# Verify pre-deployed Hyperlane7683 contracts on forked networks
verify-hyperlane: build-verify-hyperlane
	./bin/verify-hyperlane

# Deploy ERC20 tokens to all forked networks
deploy-tokens: build-deploy-tokens
	./bin/deploy-tokens

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

# Build Hyperlane7683 verification tool
build-verify-hyperlane:
	go build -o bin/verify-hyperlane ./cmd/verify-hyperlane

# Build token deployment tool
build-deploy-tokens:
	go build -o bin/deploy-tokens ./cmd/deploy-tokens

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

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
