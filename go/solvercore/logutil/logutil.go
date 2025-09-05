package logutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/NethermindEth/oif-starknet/go/solvercore/config"
)

const (
	reset     = "\033[0m"
	green     = "\033[32m"
	pastelRed = "\033[91m"
	purple    = "\033[35m"
	royalBlue = "\033[38;5;27m"
	orange    = "\033[38;5;208m"
	cyan      = "\033[36m"
)

var (
	colorByChainID map[uint64]string
	tagByChainID   map[uint64]string
	mappingOnce    sync.Once
)

func initMapping() {
	colorByChainID = make(map[uint64]string)
	tagByChainID = make(map[uint64]string)

	// 1) Bind colors/tags by env-configured chain IDs (arbitrary networks supported)
	bindEnvChain("ETHEREUM_CHAIN_ID", "[ETH]", green)
	bindEnvChain("OPTIMISM_CHAIN_ID", "[OPT]", pastelRed)
	bindEnvChain("ARBITRUM_CHAIN_ID", "[ARB]", purple)
	bindEnvChain("BASE_CHAIN_ID", "[BASE]", royalBlue)
	bindEnvChain("STARKNET_CHAIN_ID", "[STRK]", orange)

	// 2) Also bind any known configured networks by their current names
	for name, cfg := range config.Networks {
		lower := strings.ToLower(name)
		switch {
		case strings.Contains(lower, "ethereum") || strings.Contains(lower, "sepolia"):
			colorByChainID[cfg.ChainID] = green
			tagByChainID[cfg.ChainID] = "[ETH]"
		case strings.Contains(lower, "optimism") || strings.Contains(lower, "opt"):
			colorByChainID[cfg.ChainID] = pastelRed
			tagByChainID[cfg.ChainID] = "[OPT]"
		case strings.Contains(lower, "arbitrum") || strings.Contains(lower, "arb"):
			colorByChainID[cfg.ChainID] = purple
			tagByChainID[cfg.ChainID] = "[ARB]"
		case strings.Contains(lower, "base"):
			colorByChainID[cfg.ChainID] = royalBlue
			tagByChainID[cfg.ChainID] = "[BASE]"
		case strings.Contains(lower, "starknet") || strings.Contains(lower, "strk"):
			colorByChainID[cfg.ChainID] = orange
			tagByChainID[cfg.ChainID] = "[STRK]"
		}
	}
}

func bindEnvChain(key, tag, color string) {
	if v := os.Getenv(key); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			colorByChainID[id] = color
			tagByChainID[id] = tag
		}
	}
}

// Prefix returns a color-coded, short network prefix like "[ETH] " determined by env-configured chain IDs
func Prefix(networkName string) string {
	mappingOnce.Do(initMapping)
	if cfg, ok := config.Networks[networkName]; ok {
		if color, ok2 := colorByChainID[cfg.ChainID]; ok2 {
			if tag := tagByChainID[cfg.ChainID]; tag != "" {
				return fmt.Sprintf("%s%s%s ", color, tag, reset)
			}
		}
	}
	// Fallback by name substring
	if tag, color := tagColorByName(networkName); tag != "" {
		return fmt.Sprintf("%s%s%s ", color, tag, reset)
	}
	// Generic fallback
	tag := deriveTag(networkName)
	return fmt.Sprintf("%s%s%s ", cyan, tag, reset)
}

func tagColorByName(name string) (string, string) {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "ethereum") || strings.Contains(lower, "sepolia"):
		return "[ETH]", green
	case strings.Contains(lower, "optimism") || strings.Contains(lower, "opt"):
		return "[OPT]", pastelRed
	case strings.Contains(lower, "arbitrum") || strings.Contains(lower, "arb"):
		return "[ARB]", purple
	case strings.Contains(lower, "base"):
		return "[BASE]", royalBlue
	case strings.Contains(lower, "starknet") || strings.Contains(lower, "strk"):
		return "[STRK]", orange
	default:
		return "", ""
	}
}

func deriveTag(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "[NET]"
	}
	if len(parts) == 1 {
		n := parts[0]
		if len(n) > 3 {
			n = n[:3]
		}
		return fmt.Sprintf("[%s]", strings.ToUpper(n))
	}
	a := strings.ToUpper(parts[0][:1])
	b := strings.ToUpper(parts[1][:1])
	return fmt.Sprintf("[%s%s]", a, b)
}

// NetworkNameByChainID returns the first configured network name matching chainID
func NetworkNameByChainID(chainID uint64) string {
	for name, cfg := range config.Networks {
		if cfg.ChainID == chainID {
			return name
		}
	}
	return fmt.Sprintf("chain-%d", chainID)
}
