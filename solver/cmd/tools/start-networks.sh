#!/bin/bash

# Load environment variables from .env file
if [ -f ".env" ]; then
  export $(cat .env | grep -v '^#' | xargs)
  echo "📋 Loaded environment variables from .env"
fi

# Network IDs
ETHEREUM_ID="[ETH]"
OPT_ID="[OPT]"
ARB_ID="[ARB]"
BASE_ID="[BASE]"
STARKNET_ID="[STRK]"

# Colors for each network (updated color scheme)
ETHEREUM_COLOR="\033[32m"       # Green
OPT_COLOR="\033[91m"            # Pastel Red
ARB_COLOR="\033[35m"            # Purple
BASE_COLOR="\033[38;5;27m"      # Royal Blue
STARKNET_COLOR="\033[38;5;208m" # Orange
RESET="\033[0m"                 # Reset

# Function to clean up old solver state files (optional cleanup)
cleanup_old_solver_state() {
  echo "🧹 Cleaning up old solver state files..."

  # Remove old/corrupted state files in wrong locations
  rm -f "solver-state.json"
  rm -f "solvercore/config/solver-state.json"
  rm -f "solvercore/solvers/hyperlane7683/solver-state.json"

  echo "✅ Old solver state files cleaned up"
  echo "💡 Solver will create its own state file when it starts"
}

# Function to start a testnet fork with color-coded logging
start_evm_fork() {
  local port=$1
  local chain_id=$2
  local color=$3
  local id=$4
  local testnet_name=$5

  # Choose RPC endpoint based on availability
  local rpc_url
  if [ -n "$ALCHEMY_API_KEY" ]; then
    case $testnet_name in
    "sepolia")
      rpc_url="https://eth-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
      ;;
    "optimism-sepolia")
      rpc_url="https://opt-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
      ;;
    "arbitrum-sepolia")
      rpc_url="https://arb-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
      ;;
    "base-sepolia")
      rpc_url="https://base-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
      ;;
    esac
    echo -e "${color}${id}${RESET} Using Alchemy RPC for ${testnet_name}"
  else
    case $testnet_name in
    "sepolia")
      rpc_url="https://rpc.sepolia.org"
      ;;
    "optimism-sepolia")
      rpc_url="https://sepolia.optimism.io"
      ;;
    "arbitrum-sepolia")
      rpc_url="https://sepolia-rollup.arbitrum.io/rpc"
      ;;
    "base-sepolia")
      rpc_url="https://sepolia.base.org"
      ;;
    esac
    echo -e "${color}${id}${RESET} Using public RPC for ${testnet_name}"
  fi

  # Fork from latest block (no need to specify block number)
  echo -e "${color}${id}${RESET} Forking ${testnet_name} from latest block"

  # Start anvil with testnet fork and pipe output through color filter
  anvil --port $port --chain-id $chain_id --fork-url "$rpc_url" 2>&1 | while IFS= read -r line; do
    echo -e "${color}${id}${RESET} $line"
  done &

  # Store the PID
  echo $! >"/tmp/anvil_$port.pid"

  echo -e "${color}${id}${RESET} ${testnet_name} fork started on port $port (Chain ID: $chain_id)"
}

# Function to start Starknet with Katana
start_starknet_fork() {
  local port=$1
  local color=$2
  local id=$3

  echo -e "${color}${id}${RESET} Starting Starknet fork with Katana..."

  # Check if katana is installed
  if ! command -v katana &>/dev/null; then
    echo -e "${color}${id}${RESET} ❌ Katana not found. Please install it first:"
    echo -e "${color}${id}${RESET}    curl -L https://github.com/dojoengine/dojo/releases/latest/download/katana-installer.sh | bash"
    echo -e "${color}${id}${RESET}    Or visit: https://book.dojoengine.org/toolchain/katana/installation"
    return 1
  fi

  # Choose RPC endpoint based on availability
  local rpc_url
  if [ -n "$ALCHEMY_API_KEY" ]; then
    rpc_url="https://starknet-sepolia.g.alchemy.com/starknet/version/rpc/v0_8/${ALCHEMY_API_KEY}"
    echo -e "${color}${id}${RESET} Using Alchemy RPC for Starknet"
  else
    rpc_url="https://free-rpc.nethermind.io/starknet-sepolia-juno/"
    echo -e "${color}${id}${RESET} Using public RPC for Starknet"
  fi

  # Fork from latest block (no need to specify block number)
  echo -e "${color}${id}${RESET} Forking Starknet from latest block"

  katana --chain-id ${STARKNET_CHAIN_ID:-23448591} --fork.provider "$rpc_url" 2>&1 | while IFS= read -r line; do
    if [ "${LOG_LEVEL:-info}" = "debug" ]; then
      echo -e "${color}${id}${RESET} $line"
    else
      case "$line" in
      *TRACE*) ;; # skip
      *DEBUG*) ;; # skip
      *) echo -e "${color}${id}${RESET} $line" ;;
      esac
    fi
  done &

  # Store the PID
  echo $! >"/tmp/katana_$port.pid"

  echo -e "${color}${id}${RESET} Starknet fork started on port $port (Chain ID: ${STARKNET_CHAIN_ID:-23448591})"
}

# Function to validate EVM network is running and accessible
validate_evm_network() {
  local port=$1
  local network_name=$2
  local color=$3
  local id=$4
  local expected_chain_id=$5
  local pid_file="/tmp/anvil_$port.pid"
  local rpc_url="http://localhost:$port"

  # Check if PID file exists
  if [ ! -f "$pid_file" ]; then
    echo -e "${color}${id}${RESET} ❌ ${network_name} failed to start: PID file not found"
    return 1
  fi

  # Check if process is running
  local pid=$(cat "$pid_file" 2>/dev/null)
  if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
    echo -e "${color}${id}${RESET} ❌ ${network_name} failed to start: Process not running (PID: ${pid:-none})"
    return 1
  fi

  # Check if RPC endpoint is accessible
  local response
  response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
    "$rpc_url" 2>&1)

  if [ $? -ne 0 ] || [ -z "$response" ]; then
    echo -e "${color}${id}${RESET} ❌ ${network_name} RPC not accessible at $rpc_url"
    echo -e "${color}${id}${RESET}    Error: $response"
    return 1
  fi

  # Check if response contains error
  if echo "$response" | grep -q '"error"'; then
    local error_msg=$(echo "$response" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)
    echo -e "${color}${id}${RESET} ❌ ${network_name} RPC returned error: ${error_msg:-unknown error}"
    return 1
  fi

  # Verify chain ID matches (optional check)
  local chain_id_hex=$(echo "$response" | grep -o '"result":"[^"]*"' | cut -d'"' -f4)
  if [ -n "$chain_id_hex" ] && [ -n "$expected_chain_id" ]; then
    # Convert hex to decimal (remove 0x prefix if present)
    chain_id_hex=${chain_id_hex#0x}
    local chain_id_dec=$((16#$chain_id_hex))
    if [ "$chain_id_dec" != "$expected_chain_id" ]; then
      echo -e "${color}${id}${RESET} ⚠️  ${network_name} chain ID mismatch: expected $expected_chain_id, got $chain_id_dec"
    fi
  fi

  echo -e "${color}${id}${RESET} ✅ ${network_name} is running and accessible"
  return 0
}

# Function to validate Starknet network is running and accessible
validate_starknet_network() {
  local port=$1
  local network_name=$2
  local color=$3
  local id=$4
  local expected_chain_id=$5
  local pid_file="/tmp/katana_$port.pid"
  local rpc_url="http://localhost:$port"

  # Check if PID file exists
  if [ ! -f "$pid_file" ]; then
    echo -e "${color}${id}${RESET} ❌ ${network_name} failed to start: PID file not found"
    return 1
  fi

  # Check if process is running
  local pid=$(cat "$pid_file" 2>/dev/null)
  if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
    echo -e "${color}${id}${RESET} ❌ ${network_name} failed to start: Process not running (PID: ${pid:-none})"
    return 1
  fi

  # Check if RPC endpoint is accessible (Starknet uses starknet_chainId)
  local response
  response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"starknet_chainId","params":[],"id":1}' \
    "$rpc_url" 2>&1)

  if [ $? -ne 0 ] || [ -z "$response" ]; then
    echo -e "${color}${id}${RESET} ❌ ${network_name} RPC not accessible at $rpc_url"
    echo -e "${color}${id}${RESET}    Error: $response"
    return 1
  fi

  # Check if response contains error
  if echo "$response" | grep -q '"error"'; then
    local error_msg=$(echo "$response" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)
    echo -e "${color}${id}${RESET} ❌ ${network_name} RPC returned error: ${error_msg:-unknown error}"
    return 1
  fi

  echo -e "${color}${id}${RESET} ✅ ${network_name} is running and accessible"
  return 0
}

# Function to stop all networks (does not exit - caller handles exit)
cleanup_networks() {
  echo ""
  echo "🛑 Stopping all networks..."

  # Kill all anvil processes
  for port in 8545 8546 8547 8548; do
    if [ -f "/tmp/anvil_$port.pid" ]; then
      pid=$(cat "/tmp/anvil_$port.pid")
      kill $pid 2>/dev/null || true
      rm -f "/tmp/anvil_$port.pid"
    fi
  done

  # Kill Katana process
  if [ -f "/tmp/katana_5050.pid" ]; then
    pid=$(cat "/tmp/katana_5050.pid")
    kill $pid 2>/dev/null || true
    rm -f "/tmp/katana_5050.pid"
  fi

  # Also kill any remaining anvil/katana processes
  pkill -f "anvil" 2>/dev/null || true
  pkill -f "katana" 2>/dev/null || true

  echo "✅ All networks stopped"
}

# Cleanup handler for signals (exits with 0)
cleanup() {
  cleanup_networks
  exit 0
}

echo "🚀 Starting All Network Forks (EVM + Starknet)"
echo "=============================================="
echo "💡 All networks will fork from the latest block"
echo "🛑 Use Ctrl+C to stop all networks"
echo ""

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "🔧 Starting network forks..."
echo ""

# Clean up old solver state files (solver will create its own)
cleanup_old_solver_state

# Check if ALCHEMY_API_KEY is set
if [ -z "$ALCHEMY_API_KEY" ]; then
  echo "⚠️  ALCHEMY_API_KEY not set!"
  echo "💡 You'll be rate limited by the demo endpoint"
  echo "💡 Set ALCHEMY_API_KEY in your .env for more access"
  echo ""
fi

# Start all networks
# TESTING: Comment out individual lines below to test validation failures
start_evm_fork 8545 ${ETHEREUM_CHAIN_ID:-11155111} "$ETHEREUM_COLOR" "$ETHEREUM_ID" "sepolia"
start_evm_fork 8546 ${OPTIMISM_CHAIN_ID:-11155420} "$OPT_COLOR" "$OPT_ID" "optimism-sepolia"
start_evm_fork 8547 ${ARBITRUM_CHAIN_ID:-421614} "$ARB_COLOR" "$ARB_ID" "arbitrum-sepolia"
start_evm_fork 8548 ${BASE_CHAIN_ID:-84532} "$BASE_COLOR" "$BASE_ID" "base-sepolia"
start_starknet_fork 5050 "$STARKNET_COLOR" "$STARKNET_ID"

echo ""
echo "⏳ Waiting for networks to be ready..."
sleep 3

echo ""
echo "🔍 Validating network connections..."
echo ""

# Validate all networks
validation_failed=0

# Validate EVM networks
if ! validate_evm_network 8545 "Ethereum" "$ETHEREUM_COLOR" "$ETHEREUM_ID" "${ETHEREUM_CHAIN_ID:-11155111}"; then
  validation_failed=1
fi

if ! validate_evm_network 8546 "Optimism" "$OPT_COLOR" "$OPT_ID" "${OPTIMISM_CHAIN_ID:-11155420}"; then
  validation_failed=1
fi

if ! validate_evm_network 8547 "Arbitrum" "$ARB_COLOR" "$ARB_ID" "${ARBITRUM_CHAIN_ID:-421614}"; then
  validation_failed=1
fi

if ! validate_evm_network 8548 "Base" "$BASE_COLOR" "$BASE_ID" "${BASE_CHAIN_ID:-84532}"; then
  validation_failed=1
fi

# Validate Starknet network
if ! validate_starknet_network 5050 "Starknet" "$STARKNET_COLOR" "$STARKNET_ID" "${STARKNET_CHAIN_ID:-23448591}"; then
  validation_failed=1
fi

# If any validation failed, cleanup and exit
if [ $validation_failed -eq 1 ]; then
  echo ""
  echo "❌ Network validation failed. Stopping all networks..."
  cleanup_networks
  exit 1
fi

echo ""
echo "🎉 All network forks are running!"
echo "================================"
echo -e "${ETHEREUM_COLOR}${ETHEREUM_ID}${RESET} Ethereum Fork    - http://localhost:8545 (Chain ID: ${ETHEREUM_CHAIN_ID:-11155111})"
echo -e "${OPT_COLOR}${OPT_ID}${RESET} Optimism Fork    - http://localhost:8546 (Chain ID: ${OPTIMISM_CHAIN_ID:-11155420})"
echo -e "${ARB_COLOR}${ARB_ID}${RESET} Arbitrum Fork    - http://localhost:8547 (Chain ID: ${ARBITRUM_CHAIN_ID:-421614})"
echo -e "${BASE_COLOR}${BASE_ID}${RESET} Base Fork       - http://localhost:8548 (Chain ID: ${BASE_CHAIN_ID:-84532})"
echo -e "${STARKNET_COLOR}${STARKNET_ID}${RESET} Starknet Fork   - http://localhost:5050 (Chain ID: ${STARKNET_CHAIN_ID:-23448591})"
echo ""
echo "💡 Networks will continue logging here..."
echo "💡 Solver will create its own state file when it starts"
echo "🛑 Press Ctrl+C to stop all networks"
echo ""

# Wait for user to stop
echo "⏳ Networks running... (Press Ctrl+C to stop)"
wait
