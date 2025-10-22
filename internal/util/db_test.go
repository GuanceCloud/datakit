// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestObfuscateSQL(t *testing.T) {
	sql := "select *    \n from table where id = 1"
	obfuscated := ObfuscateSQL(sql)
	assert.Equal(t, "select * from table where id = ?", obfuscated)

	// invalid sql
	sql = "select * from table where id='invalid"
	obfuscated = ObfuscateSQL(sql)
	assert.Contains(t, obfuscated, "ERROR:")
}

func TestObfuscateSQLExecPlan(t *testing.T) {
	plan := `{
		"Plan": {
			"Node Type": "Seq Scan",
			"Filter": "id = 1"
		}
	}`
	obfuscated := ObfuscateSQLExecPlan(plan, &ObfuscateLogger{})
	assert.Equal(t, `{"Plan":{"Node Type":"Seq Scan","Filter":"id = ?"}}`, obfuscated)
}

func TestComputeSQLSignature(t *testing.T) {
	sql := "select * from table where id = 1"
	signature := ComputeSQLSignature(sql)
	assert.Equal(t, "564d85a38e9adf2e3523a8f90cee50ca", signature)
}

func TestCacheLimit(t *testing.T) {
	t.Run("acquire", func(t *testing.T) {
		tests := []struct {
			name    string
			size    int
			ttl     int64
			key     string
			setup   func(*CacheLimit)
			want    bool
			wantLen int
		}{{
			name:    "acquire new key when cache is empty",
			size:    2,
			ttl:     10,
			key:     "testKey",
			setup:   func(cl *CacheLimit) {},
			want:    true,
			wantLen: 1,
		}, {
			name: "acquire existing non-expired key",
			size: 2,
			ttl:  10,
			key:  "testKey",
			setup: func(cl *CacheLimit) {
				cl.Acquire("testKey")
			},
			want:    false,
			wantLen: 1,
		}, {
			name: "acquire when cache size limit reached",
			size: 1,
			ttl:  10,
			key:  "testKey2",
			setup: func(cl *CacheLimit) {
				cl.Acquire("testKey1")
			},
			want:    false,
			wantLen: 1,
		}, {
			name: "acquire expired key",
			size: 1,
			ttl:  1,
			key:  "testKey",
			setup: func(cl *CacheLimit) {
				cl.Acquire("testKey")
				time.Sleep(2 * time.Second) // Wait for TTL expiration
			},
			want:    true,
			wantLen: 1,
		}}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cl := &CacheLimit{
					Size: tt.size,
					TTL:  tt.ttl,
				}
				tt.setup(cl)
				got := cl.Acquire(tt.key)
				if got != tt.want {
					t.Errorf("Acquire() = %v, want %v", got, tt.want)
				}
				if l := cl.len(); l != tt.wantLen {
					t.Errorf("len() = %v, want %v", l, tt.wantLen)
				}
			})
		}
	})

	t.Run("len", func(t *testing.T) {
		cl := &CacheLimit{
			TTL: 10,
			itemStore: map[string]cacheItem{
				"validKey":   {expire: time.Now().Add(5 * time.Second)},
				"expiredKey": {expire: time.Now().Add(-5 * time.Second)},
			},
		}

		assert.Equal(t, 1, cl.len())
	})

	t.Run("get", func(t *testing.T) {
		cl := &CacheLimit{
			TTL: 10,
			itemStore: map[string]cacheItem{
				"validKey":   {expire: time.Now().Add(5 * time.Second)},
				"expiredKey": {expire: time.Now().Add(-5 * time.Second)},
			},
		}

		// Test valid key
		if !cl.get("validKey") {
			t.Error("get(validKey) should return true for non-expired key")
		}

		// Test expired key
		if cl.get("expiredKey") {
			t.Error("get(expiredKey) should return false for expired key")
		}

		// Verify expired key was deleted
		if _, exists := cl.itemStore["expiredKey"]; exists {
			t.Error("get() should delete expired keys from itemStore")
		}
	})

	t.Run("add", func(t *testing.T) {
		cl := &CacheLimit{
			Size: 1,
			TTL:  1,
		}

		cl.add("testKey")

		// Verify item was added with correct expiration
		if item, exists := cl.itemStore["testKey"]; !exists {
			t.Error("add() should store new key in itemStore")
		} else if item.expire.Before(time.Now()) {
			t.Error("add() should set correct expiration time")
		}
	})
}
