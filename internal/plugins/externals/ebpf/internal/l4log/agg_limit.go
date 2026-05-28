//go:build linux
// +build linux

package l4log

import (
	"os"
	"strconv"
	"strings"
)

const (
	l4logNetflowAggLimitEnv     = "DK_EBPF_L4LOG_NETFLOW_AGG_LIMIT"
	l4logHTTPAggLimitEnv        = "DK_EBPF_L4LOG_HTTP_AGG_LIMIT"
	defaultL4logNetflowAggLimit = 32_768
	defaultL4logHTTPAggLimit    = 32_768
	maxL4logAggLimit            = 1_000_000
)

func l4logAggLimitFromEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		log.Warnf("invalid %s=%q, use default %d", key, raw, fallback)
		return fallback
	}
	if limit > maxL4logAggLimit {
		return maxL4logAggLimit
	}
	return limit
}

func (agg *FlowAggTCP) entryLimit() int {
	if agg == nil {
		return defaultL4logNetflowAggLimit
	}
	if agg.limit <= 0 {
		agg.limit = l4logAggLimitFromEnv(l4logNetflowAggLimitEnv, defaultL4logNetflowAggLimit)
	}
	return agg.limit
}

func (agg *FlowAggHTTP) entryLimit() int {
	if agg == nil {
		return defaultL4logHTTPAggLimit
	}
	if agg.limit <= 0 {
		agg.limit = l4logAggLimitFromEnv(l4logHTTPAggLimitEnv, defaultL4logHTTPAggLimit)
	}
	return agg.limit
}
