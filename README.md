# OIF-Starknet 🚀

Extension of BootNodeDev's [Open Intent Framework](https://github.com/BootNodeDev/intents-framework/tree/main) for adding Starknet support.

## Overview

This project extends the Open Intent Framework to support cross-chain order filling between EVM chains and Starknet using the Hyperlane protocol.

## Project Structure

- `solidity/`: Copy of [Open Intent Framework](https://github.com/BootNodeDev/intents-framework/tree/main) Solidity contracts
- `cairo/`: Starknet translation of Solidity contracts  
- `solver/`: Hyperlane intent solver implementation - [See detailed README](solver/README.md)

## Documentation

- **Go Solver Implementation**: [solver/README.md](solver/README.md) - Detailed architecture, setup, and development guide
- **Testing Commands**: Run `cd solver/ && make help` for complete command reference

## License

Apache-2.0
