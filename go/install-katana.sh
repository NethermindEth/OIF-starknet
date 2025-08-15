#!/bin/bash

echo "🔧 Installing Katana for Starknet development..."

# Check if katana is already installed
if command -v katana &> /dev/null; then
    echo "✅ Katana is already installed!"
    katana --version
    exit 0
fi

echo "📥 Installing Katana..."

# Try the official installer first
if curl -L https://github.com/dojoengine/dojo/releases/latest/download/katana-installer.sh | bash; then
    echo "✅ Katana installed successfully via official installer!"
    
    # Source the profile to make katana available in current session
    if [ -f "$HOME/.bashrc" ]; then
        source "$HOME/.bashrc"
    elif [ -f "$HOME/.zshrc" ]; then
        source "$HOME/.zshrc"
    fi
    
    # Verify installation
    if command -v katana &> /dev/null; then
        echo "✅ Katana is now available!"
        katana --version
    else
        echo "⚠️  Katana installed but not in PATH. Please restart your terminal or run:"
        echo "   source ~/.bashrc  # or source ~/.zshrc"
    fi
else
    echo "❌ Official installer failed. Trying alternative methods..."
    
    # Alternative: Check if we can use cargo
    if command -v cargo &> /dev/null; then
        echo "📦 Installing via Cargo..."
        if cargo install --git https://github.com/dojoengine/dojo --bin katana; then
            echo "✅ Katana installed successfully via Cargo!"
            katana --version
        else
            echo "❌ Cargo installation also failed."
        fi
    else
        echo "❌ Neither installer nor Cargo worked."
        echo ""
        echo "💡 Manual installation options:"
        echo "   1. Visit: https://book.dojoengine.org/toolchain/katana/installation"
        echo "   2. Download from: https://github.com/dojoengine/dojo/releases"
        echo "   3. Install Rust and use: cargo install --git https://github.com/dojoengine/dojo --bin katana"
    fi
fi

echo ""
echo "🚀 After installation, you can start all networks with:"
echo "   ./start-networks.sh"
