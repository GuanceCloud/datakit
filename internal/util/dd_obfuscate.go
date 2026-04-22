// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package util

import (
	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

var (
	// defaultSQLPlanNormalizeSettings are the default JSON obfuscator settings for both obfuscating and normalizing SQL
	// execution plans.
	ddDefaultSQLPlanNormalizeSettings = obfuscate.JSONConfig{
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

	// defaultSQLPlanObfuscateSettings builds upon sqlPlanNormalizeSettings by including cost & row estimates in the keep
	// list.
	ddDefaultSQLPlanObfuscateSettings = obfuscate.JSONConfig{
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
		}, ddDefaultSQLPlanNormalizeSettings.KeepValues...),
		ObfuscateSQLValues: defaultSQLPlanNormalizeSettings.ObfuscateSQLValues,
	}
)

// NewSQLPlanObfuscator creates a new Obfuscator for execution plan.
func NewSQLPlanObfuscator() *obfuscate.Obfuscator {
	cfg := obfuscate.Config{}
	if !cfg.SQLExecPlan.Enabled {
		cfg.SQLExecPlan = ddDefaultSQLPlanObfuscateSettings
	}
	if !cfg.SQLExecPlanNormalize.Enabled {
		cfg.SQLExecPlanNormalize = ddDefaultSQLPlanNormalizeSettings
	}
	return obfuscate.NewObfuscator(cfg)
}
