#!/bin/bash

# Load environment variables from .env file
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "📋 Loaded environment variables from .env"
fi

# Colors for each network
ETH_COLOR="\033[36m"      # Cyan
OPT_COLOR="\033[32m"      # Green
ARB_COLOR="\033[33m"      # Yellow
BASE_COLOR="\033[35m"     # Magenta
RESET="\033[0m"           # Reset

# Network IDs
ETH_ID="[ETH]"
OPT_ID="[OPT]"
ARB_ID="[ARB]"
BASE_ID="[BASE]"

echo "🚀 Starting All Mainnet Forks"
echo "=============================="
echo "💡 All networks will fork mainnet with real infrastructure"
echo "🛑 Use Ctrl+C to stop all networks"
echo ""

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
    
    # Fork from latest block (testnets are smaller, so no need for specific block)
    echo -e "${color}${id}${RESET} Forking ${testnet_name} from latest block"
    
    # Start anvil with testnet fork and pipe output through color filter
    anvil --port $port --chain-id $chain_id --fork-url "$rpc_url" 2>&1 | while IFS= read -r line; do
        echo -e "${color}${id}${RESET} $line"
    done &
    
    # Store the PID
    echo $! > "/tmp/anvil_$port.pid"
    
    echo -e "${color}${id}${RESET} ${testnet_name} fork started on port $port (Chain ID: $chain_id)"
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
    
    # Also kill any remaining anvil processes
    pkill -f "anvil" 2>/dev/null || true
    
    echo "✅ All networks stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "🔧 Starting testnet forks..."
echo ""

# Check if ALCHEMY_API_KEY is set
if [ -z "$ALCHEMY_API_KEY" ]; then
    echo "⚠️  ALCHEMY_API_KEY not set!"
    echo "💡 You'll be rate limited by the demo endpoint"
    echo "💡 Set ALCHEMY_API_KEY in your .env for full access"
    echo "💡 Or use alternative RPC endpoints (see script for options)"
    echo ""
    echo "🔗 Alternative RPC endpoints (free tiers):"
    echo "   • Sepolia: https://rpc.sepolia.org"
    echo "   • Optimism Sepolia: https://sepolia.optimism.io"
    echo "   • Arbitrum Sepolia: https://sepolia-rollup.arbitrum.io/rpc"
    echo "   • Base Sepolia: https://sepolia.base.org"
    echo ""
fi

# Start all networks
start_network 8545 31337 "$ETH_COLOR" "$ETH_ID" "sepolia"
start_network 8546 31338 "$OPT_COLOR" "$OPT_ID" "optimism-sepolia"
start_network 8547 31339 "$ARB_COLOR" "$ARB_ID" "arbitrum-sepolia"
start_network 8548 31340 "$BASE_COLOR" "$BASE_ID" "base-sepolia"

echo ""
echo "⏳ Waiting for networks to be ready..."
sleep 3

echo ""
echo "🎉 All testnet forks are running!"
echo "================================"
echo -e "${ETH_COLOR}${ETH_ID}${RESET} Sepolia Fork             - http://localhost:8545 (Chain ID: 31337)"
echo -e "${OPT_COLOR}${OPT_ID}${RESET} Optimism Sepolia Fork    - http://localhost:8546 (Chain ID: 31338)"
echo -e "${ARB_COLOR}${ARB_ID}${RESET} Arbitrum Sepolia Fork    - http://localhost:8547 (Chain ID: 31339)"
echo -e "${BASE_COLOR}${BASE_ID}${RESET} Base Sepolia Fork        - http://localhost:8548 (Chain ID: 31340)"
echo ""
echo "🚀 What you get for FREE on all forks:"
echo "   • Permit2 at 0x000000000022D473030F116dDEE9F6B43aC78BA3"
echo "   • USDC, WETH, and other real tokens"
echo "   • Hyperlane Mailbox and infrastructure"
echo "   • Real gas dynamics and market conditions"
echo ""
echo "📦 Next steps:"
echo "   1. Fund accounts: make fund-accounts"
echo "   2. Deploy Hyperlane7683: make deploy-hyperlane"
echo ""
echo "🔄 Or restart everything:"
echo "   make restart"
echo ""
echo "💡 Networks will continue logging here..."
echo "🛑 Press Ctrl+C to stop all networks"
echo ""

# Wait for user to stop
echo "⏳ Networks running... (Press Ctrl+C to stop)"
wait
