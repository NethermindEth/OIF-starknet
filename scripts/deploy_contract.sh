#!/bin/bash

# Ensure the script stops on the first error
set -e

# Load environment variables from .env file
if [ -f .env ]; then
	export $(grep -v '^#' .env | xargs)
else
	echo ".env file not found. Please create one based on env.example."
	exit 1
fi

# Check required environment variables
if [ -z "$STARKNET_RPC" ] || [ -z "$STARKNET_ACCOUNT" ] || [ -z "$STARKNET_PRIVATE_KEY" ]; then
	echo "ERROR - One or more required environment variables are missing."
	exit 1
fi

# Set environment for starkli
export STARKNET_RPC
export STARKNET_ACCOUNT
export STARKNET_PRIVATE_KEY

# Declare contract
echo
echo "=========================="
echo "Running scarb build..."
echo "=========================="
echo
#cd ../cairo && scarb build && cd ../scripts

# Paths
CONTRACT_NAME="Hyperlane7683"
CONTRACT_JSON="../cairo/target/dev/oif_starknet_${CONTRACT_NAME}.contract_class.json"

# Declare contract
echo
echo "=========================="
echo "Declare $CONTRACT_NAME"
echo "=========================="
echo
CLASS_HASH=$(starkli declare "$CONTRACT_JSON" --watch | grep -o '0x[a-fA-F0-9]\{64\}' | head -1)
echo "[$CONTRACT_NAME] Class hash declared: $CLASS_HASH"

# Deploy contract
echo
echo "=========================="
echo "Deploy $CONTRACT_NAME"
echo "=========================="
echo

# Constructor arguments in order:
# 1. permit2: ContractAddress
# 2. mailbox: ContractAddress
# 3. owner: ContractAddress
# 4. hook: ContractAddress
# 5. interchain_security_module: ContractAddress

# Use the environment variables that are already set up
PERMIT2_ADDRESS="0x02286537be3743c9cce6fc9a442cb025c8cae688a671462b732a24d4ffa54889"
MAILBOX_ADDRESS="0x03c725cd6a4463e4a9258d29304bcca5e4f1bbccab078ffd69784f5193a6d792"
OWNER_ADDRESS="0x3ff18229c6066a59e997e5f2164b15a6bf26c9de78d755244ca2d2a678e981b"
HOOK_ADDRESS="0x1eff3a364cb5ec3ebef9267d0cc3ebcb22cb983af981d7c128fc8bad30b6bc2"
ISM_ADDRESS="0x5c4b276e622a419c59da565197f200bdca4a5fb26dcb85a45cfa9ea66958ebb"

echo "Constructor arguments:"
echo "  permit2: $PERMIT2_ADDRESS"
echo "  mailbox: $MAILBOX_ADDRESS"
echo "  owner: $OWNER_ADDRESS"
echo "  hook: $HOOK_ADDRESS"
echo "  interchain_security_module: $ISM_ADDRESS"

CONTRACT_ADDRESS=$(starkli deploy "$CLASS_HASH" \
	"$PERMIT2_ADDRESS" \
	"$MAILBOX_ADDRESS" \
	"$OWNER_ADDRESS" \
	"$HOOK_ADDRESS" \
	"$ISM_ADDRESS" \
	--watch | grep -o '0x[a-fA-F0-9]\{64\}' | head -1)
echo "[$CONTRACT_NAME] Contract deployed at: $CONTRACT_ADDRESS"

# Save deployment details to latest_deployment.txt
echo -e "$CONTRACT_NAME address: $CONTRACT_ADDRESS\n\n$CONTRACT_NAME class hash: $CLASS_HASH" >latest_deployment.txt
echo -e "Deployment address & class hash saved to latest_deployment.txt\n$CONTRACT_NAME address: $CONTRACT_ADDRESS\n$CONTRACT_NAME class hash: $CLASS_HASH"
