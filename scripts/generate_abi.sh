#!/bin/bash

# Add contracts here as "contract_name:abi_name"
contracts=(
	"Permit2:permit2"
)

echo "Running scarb build..."
cd ../oif_starknet/ && scarb build && cd ../scripts

# Generate ABIs
for contract in "${contracts[@]}"; do
	IFS=':' read -r contract_name abi_name <<<"$contract"
	json_file="../oif_starknet/target/dev/oif_starknet_${contract_name}.contract_class.json"
	abi_file="../ABI/${abi_name}.ts"

	npx abi-wan-kanabi --input "$json_file" --output "$abi_file"
done
