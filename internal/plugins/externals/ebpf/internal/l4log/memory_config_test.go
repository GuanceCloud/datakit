//go:build linux
// +build linux

package l4log

import (
	"sync"
	"testing"
)

func TestNetlogBaseKVsCacheDefaultEnabled(t *testing.T) {
	t.Setenv(envNetlogBaseKVsCache, "")
	netlogMemoryConfigOnce = sync.Once{}

	if !netlogBaseKVsCacheEnabled() {
		t.Fatal("expected base KVs cache to be enabled by default")
	}
}

func TestNetlogBaseKVsCacheCanBeDisabled(t *testing.T) {
	t.Setenv(envNetlogBaseKVsCache, "0")
	netlogMemoryConfigOnce = sync.Once{}

	if netlogBaseKVsCacheEnabled() {
		t.Fatal("expected base KVs cache to be disabled by env")
	}
}
