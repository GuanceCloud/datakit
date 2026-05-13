//go:build linux
// +build linux

package l4log

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	envNetlogBaseKVsCache     = "DKE_NETLOG_BASE_KV_CACHE"
	envNetlogMaxChunksPerConn = "DKE_NETLOG_MAX_CHUNKS_PER_CONN"

	defaultNetlogMaxChunksPerConn = 32
)

type netlogMemoryConfig struct {
	baseKVsCache     bool
	maxChunksPerConn int
}

var (
	netlogMemoryConfigOnce sync.Once
	netlogMemoryCfg        netlogMemoryConfig
)

func loadNetlogMemoryConfig() {
	netlogMemoryCfg = netlogMemoryConfig{
		baseKVsCache:     parseBoolEnv(envNetlogBaseKVsCache, true),
		maxChunksPerConn: defaultNetlogMaxChunksPerConn,
	}

	if raw := strings.TrimSpace(os.Getenv(envNetlogMaxChunksPerConn)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			log.Warnf("invalid %s=%q: %v", envNetlogMaxChunksPerConn, raw, err)
			return
		}
		if n < 0 {
			log.Warnf("invalid %s=%q: must be >= 0", envNetlogMaxChunksPerConn, raw)
			return
		}
		netlogMemoryCfg.maxChunksPerConn = n
	}
}

func currentNetlogMemoryConfig() netlogMemoryConfig {
	netlogMemoryConfigOnce.Do(loadNetlogMemoryConfig)
	return netlogMemoryCfg
}

func netlogBaseKVsCacheEnabled() bool {
	return currentNetlogMemoryConfig().baseKVsCache
}

func netlogMaxChunksPerConn() int {
	return currentNetlogMemoryConfig().maxChunksPerConn
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback
	}

	switch raw {
	case "1", "t", "true", "y", "yes", "on", "enable", "enabled":
		return true
	case "0", "f", "false", "n", "no", "off", "disable", "disabled":
		return false
	default:
		log.Warnf("invalid %s=%q, use default %t", key, raw, fallback)
		return fallback
	}
}
