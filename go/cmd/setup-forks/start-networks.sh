#!/bin/bash

# Load environment variables from .env file
if [ -f ".env" ]; then
	export $(cat .env | grep -v '^#' | xargs)
	echo "📋 Loaded environment variables from .env"
fi

# Colors for each network (updated color scheme)
ETHEREUM_COLOR="\033[32m"       # Green
OPT_COLOR="\033[91m"            # Pastel Red
ARB_COLOR="\033[35m"            # Purple
BASE_COLOR="\033[38;5;27m"      # Royal Blue
STARKNET_COLOR="\033[38;5;208m" # Orange
RESET="\033[0m"                 # Reset

# Network IDs
ETHEREUM_ID="[ETH]"
OPT_ID="[OPT]"
ARB_ID="[ARB]"
BASE_ID="[BASE]"
STARKNET_ID="[STRK]"

echo "🚀 Starting All Network Forks (EVM + Starknet)"
echo "=============================================="
echo "💡 All networks will fork mainnet with real infrastructure"
echo "🛑 Use Ctrl+C to stop all networks"
echo ""

# Function to reset solver state to fork block numbers
reset_solver_state() {
	echo "🔄 Resetting solver state to solver start block numbers..."

	# Ensure solver state directory exists (local go/state/solver_state)
	STATE_DIR="state/solver_state"
	mkdir -p "$STATE_DIR"
	STATE_FILE="$STATE_DIR/solver-state.json"

	# Get values from environment variables with defaults
	ETHEREUM_SOLVER_BLOCK=${ETHEREUM_SOLVER_START_BLOCK:-8319000}
	OPTIMISM_SOLVER_BLOCK=${OPTIMISM_SOLVER_START_BLOCK:-27370000}
	ARBITRUM_SOLVER_BLOCK=${ARBITRUM_SOLVER_START_BLOCK:-138020000}
	BASE_SOLVER_BLOCK=${BASE_SOLVER_START_BLOCK:-25380000}
	STARKNET_SOLVER_BLOCK=${STARKNET_SOLVER_START_BLOCK:-1770800}

	# Create the solver state JSON with only last indexed blocks
	# Note: All addresses and chain IDs now come from .env file via config package
	cat >"$STATE_FILE" <<EOF
{
  "networks": {
    "Ethereum": {
      "lastIndexedBlock": ${ETHEREUM_SOLVER_BLOCK},
      "lastUpdated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    },
    "Optimism": {
      "lastIndexedBlock": ${OPTIMISM_SOLVER_BLOCK},
      "lastUpdated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    },
    "Arbitrum": {
      "lastIndexedBlock": ${ARBITRUM_SOLVER_BLOCK},
      "lastUpdated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    },
    "Base": {
      "lastIndexedBlock": ${BASE_SOLVER_BLOCK},
      "lastUpdated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    },
    "Starknet": {
      "lastIndexedBlock": ${STARKNET_SOLVER_BLOCK},
      "lastUpdated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    }
  }
}
EOF

	echo "✅ Solver state reset to solver start block numbers"
	echo "🔧 Using environment variables for solver start blocks:"
	echo "   • ETHEREUM_SOLVER_START_BLOCK: ${ETHEREUM_SOLVER_BLOCK}"
	echo "   • OPTIMISM_SOLVER_START_BLOCK: ${OPTIMISM_SOLVER_BLOCK}"
	echo "   • ARBITRUM_SOLVER_START_BLOCK: ${ARBITRUM_SOLVER_BLOCK}"
	echo "   • BASE_SOLVER_START_BLOCK: ${BASE_SOLVER_BLOCK}"
	echo "   • STARKNET_SOLVER_START_BLOCK: ${STARKNET_SOLVER_BLOCK}"
	echo "💡 All contract addresses and config come from .env file"
}

# Function to start a testnet fork with color-coded logging
start_network() {
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

	# Fork from block when contract was last active to preserve state
	local fork_block
	case $testnet_name in
	"sepolia")
		fork_block=${ETHEREUM_SOLVER_START_BLOCK:-8319000} # SolverStartBlock - 1
		;;
	"optimism-sepolia")
		fork_block=${OPTIMISM_SOLVER_START_BLOCK:-27370000} # SolverStartBlock - 1
		;;
	"arbitrum-sepolia")
		fork_block=${ARBITRUM_SOLVER_START_BLOCK:-138020000} # SolverStartBlock - 1
		;;
	"base-sepolia")
		fork_block=${BASE_SOLVER_START_BLOCK:-25380000} # SolverStartBlock - 1
		;;
	*)
		fork_block=${ETHEREUM_SOLVER_START_BLOCK:-8319000}
		;;
	esac
	echo -e "${color}${id}${RESET} Forking ${testnet_name} from block ${fork_block} (SolverStartBlock - 1)"

	# Start anvil with testnet fork and pipe output through color filter
	anvil --port $port --chain-id $chain_id --fork-url "$rpc_url" --fork-block-number ${fork_block} 2>&1 | while IFS= read -r line; do
		echo -e "${color}${id}${RESET} $line"
	done &

	# Store the PID
	echo $! >"/tmp/anvil_$port.pid"

	echo -e "${color}${id}${RESET} ${testnet_name} fork started on port $port (Chain ID: $chain_id)"
}

# Function to start Starknet with Katana
start_starknet() {
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

	# Start Katana with state forking (SolverStartBlock - 1)
	local fork_block=${STARKNET_SOLVER_START_BLOCK:-1530000}
	katana --chain-id $STARKNET_SOLVER_START_BLOCK --fork.provider "$rpc_url" --fork.block ${fork_block} 2>&1 | while IFS= read -r line; do
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

	echo -e "${color}${id}${RESET} Starknet fork started on port $port (Chain ID: 23448591)"
}

# Function to stop all networks
cleanup() {
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
	exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "🔧 Starting network forks..."
echo ""

# Reset solver state to fork block numbers
reset_solver_state

# Check if ALCHEMY_API_KEY is set
if [ -z "$ALCHEMY_API_KEY" ]; then
	echo "⚠️  ALCHEMY_API_KEY not set!"
	echo "💡 You'll be rate limited by the demo endpoint"
	echo "💡 Set ALCHEMY_API_KEY in your .env for full access"
	echo ""
fi

# Start all networks
start_network 8545 11155111 "$ETHEREUM_COLOR" "$ETHEREUM_ID" "sepolia"
start_network 8546 11155420 "$OPT_COLOR" "$OPT_ID" "optimism-sepolia"
start_network 8547 421614 "$ARB_COLOR" "$ARB_ID" "arbitrum-sepolia"
start_network 8548 84532 "$BASE_COLOR" "$BASE_ID" "base-sepolia"
start_starknet 5050 "$STARKNET_COLOR" "$STARKNET_ID"

echo ""
echo "⏳ Waiting for networks to be ready..."
sleep 5

echo ""
echo "🎉 All network forks are running!"
echo "================================"
echo -e "${ETHEREUM_COLOR}${ETHEREUM_ID}${RESET} Ethereum Fork  - http://localhost:8545 (Chain ID: 11155111)"
echo -e "${OPT_COLOR}${OPT_ID}${RESET} Optimism Fork            - http://localhost:8546 (Chain ID: 11155420)"
echo -e "${ARB_COLOR}${ARB_ID}${RESET} Arbitrum Fork            - http://localhost:8547 (Chain ID: 421614)"
echo -e "${BASE_COLOR}${BASE_ID}${RESET} Base Fork              - http://localhost:8548 (Chain ID: 84532)"
echo -e "${STARKNET_COLOR}${STARKNET_ID}${RESET} Starknet Fork  - http://localhost:5050 (Chain ID: 23448591)"
echo ""
echo "💡 Networks will continue logging here..."
echo "💡 Solver will automatically start from solver state blocks"
echo "🛑 Press Ctrl+C to stop all networks"
echo ""

# Wait for user to stop
echo "⏳ Networks running... (Press Ctrl+C to stop)"
wait
