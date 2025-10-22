// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package util

import (
	"crypto/md5" //nolint:gosec
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
)

var reg = regexp.MustCompile(`\n|\s+`) //nolint:gocritic

func ObfuscateSQL(text string) (sql string) {
	defer func() {
		sql = strings.TrimSpace(reg.ReplaceAllString(sql, " "))
	}()

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", text); err != nil {
		return fmt.Sprintf("ERROR: failed to obfuscate: %s", err.Error())
	} else {
		return out.Query
	}
}

var defaultSQLPlanNormalizeSettings = obfuscate.JSONConfig{
	Enabled: true,
	ObfuscateSQLValues: []string{
		// mysql
		"attached_condition",
		// postgres
		"Cache Key",
		"Conflict Filter",
		"Function Call",
		"Filter",
		"Hash Cond",
		"Index Cond",
		"Join Filter",
		"Merge Cond",
		"Output",
		"Recheck Cond",
		"Repeatable Seed",
		"Sampling Parameters",
		"TID Cond",
	},
	KeepValues: []string{
		// mysql
		"access_type",
		"backward_index_scan",
		"cacheable",
		"delete",
		"dependent",
		"first_match",
		"key",
		"key_length",
		"possible_keys",
		"ref",
		"select_id",
		"table_name",
		"update",
		"used_columns",
		"used_key_parts",
		"using_MRR",
		"using_filesort",
		"using_index",
		"using_join_buffer",
		"using_temporary_table",
		// postgres
		"Actual Loops",
		"Actual Rows",
		"Actual Startup Time",
		"Actual Total Time",
		"Alias",
		"Async Capable",
		"Average Sort Space Used",
		"Cache Evictions",
		"Cache Hits",
		"Cache Misses",
		"Cache Overflows",
		"Calls",
		"Command",
		"Conflict Arbiter Indexes",
		"Conflict Resolution",
		"Conflicting Tuples",
		"Constraint Name",
		"CTE Name",
		"Custom Plan Provider",
		"Deforming",
		"Emission",
		"Exact Heap Blocks",
		"Execution Time",
		"Expressions",
		"Foreign Delete",
		"Foreign Insert",
		"Foreign Update",
		"Full-sort Group",
		"Function Name",
		"Generation",
		"Group Count",
		"Grouping Sets",
		"Group Key",
		"HashAgg Batches",
		"Hash Batches",
		"Hash Buckets",
		"Heap Fetches",
		"I/O Read Time",
		"I/O Write Time",
		"Index Name",
		"Inlining",
		"Join Type",
		"Local Dirtied Blocks",
		"Local Hit Blocks",
		"Local Read Blocks",
		"Local Written Blocks",
		"Lossy Heap Blocks",
		"Node Type",
		"Optimization",
		"Original Hash Batches",
		"Original Hash Buckets",
		"Parallel Aware",
		"Parent Relationship",
		"Partial Mode",
		"Peak Memory Usage",
		"Peak Sort Space Used",
		"Planned Partitions",
		"Planning Time",
		"Pre-sorted Groups",
		"Presorted Key",
		"Query Identifier",
		"Relation Name",
		"Rows Removed by Conflict Filter",
		"Rows Removed by Filter",
		"Rows Removed by Index Recheck",
		"Rows Removed by Join Filter",
		"Sampling Method",
		"Scan Direction",
		"Schema",
		"Settings",
		"Shared Dirtied Blocks",
		"Shared Hit Blocks",
		"Shared Read Blocks",
		"Shared Written Blocks",
		"Single Copy",
		"Sort Key",
		"Sort Method",
		"Sort Methods Used",
		"Sort Space Type",
		"Sort Space Used",
		"Strategy",
		"Subplan Name",
		"Subplans Removed",
		"Target Tables",
		"Temp Read Blocks",
		"Temp Written Blocks",
		"Time",
		"Timing",
		"Total",
		"Trigger",
		"Trigger Name",
		"Triggers",
		"Tuples Inserted",
		"Tuplestore Name",
		"WAL Bytes",
		"WAL FPI",
		"WAL Records",
		"Worker",
		"Worker Number",
		"Workers",
		"Workers Launched",
		"Workers Planned",
	},
}

type ObfuscateLogger struct {
	Log *logger.Logger
}

func (l *ObfuscateLogger) Errorf(s string, args ...interface{}) error {
	if l.Log == nil {
		return nil
	}
	l.Log.Errorf(s, args...)
	return nil
}

func (l *ObfuscateLogger) Debugf(s string, args ...interface{}) {
	if l.Log == nil {
		return
	}
	l.Log.Debugf(s, args...)
}

func ObfuscateSQLExecPlan(plan string, log *ObfuscateLogger) (obfPlan string) {
	obfPlan, err := obfuscate.NewObfuscator(&obfuscate.Config{
		Log: log,
		SQLExecPlan: obfuscate.JSONConfig{
			Enabled: true,
			KeepValues: append([]string{
				// mysql
				"cost_info",
				"filtered",
				"rows_examined_per_join",
				"rows_examined_per_scan",
				"rows_produced_per_join",
				// postgres
				"Plan Rows",
				"Plan Width",
				"Startup Cost",
				"Total Cost",
			}, defaultSQLPlanNormalizeSettings.KeepValues...),
			ObfuscateSQLValues: defaultSQLPlanNormalizeSettings.ObfuscateSQLValues,
		},
		SQLExecPlanNormalize: obfuscate.JSONConfig{
			Enabled: true,
		},
	}).ObfuscateSQLExecPlan(plan, false)
	if err != nil {
		return fmt.Sprintf("ERROR: failed to obfuscate: %s", err.Error())
	}
	return
}

func ComputeSQLSignature(text string) (signature string) {
	signature = fmt.Sprintf("%x", md5.Sum([]byte(text))) //nolint:gosec
	return
}

type cacheItem struct {
	expire time.Time
}

// CacheLimit is a simple ttl cache.
type CacheLimit struct {
	Size      int                  // max cache size
	TTL       int64                // max time in seconds to live
	itemStore map[string]cacheItem // cache items
}

func (c *CacheLimit) len() int {
	count := 0
	for k := range c.itemStore {
		if ok := c.get(k); ok {
			count++
		}
	}

	return count
}

func (c *CacheLimit) get(key string) bool {
	if value, ok := c.itemStore[key]; ok {
		if time.Now().Before(value.expire) {
			return true
		} else {
			delete(c.itemStore, key)
			return false
		}
	}
	return false
}

func (c *CacheLimit) add(key string) {
	if c.itemStore == nil {
		c.itemStore = make(map[string]cacheItem)
	}
	duration := time.Duration(c.TTL) * time.Second
	expire := time.Now().Add(duration)
	c.itemStore[key] = cacheItem{
		expire: expire,
	}
}

func (c *CacheLimit) Acquire(key string) bool {
	if c.len() >= c.Size {
		return false
	}
	if ok := c.get(key); !ok {
		c.add(key)
		return true
	}
	return false
}
