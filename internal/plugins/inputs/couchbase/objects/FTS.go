// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	// Sample Keys.
	FtsCurrBatchesBlockedByHerder   = "fts_curr_batches_blocked_by_herder"
	FtsNumBytesUsedRAM              = "fts_num_bytes_used_ram"
	FtsTotalQueriesRejectedByHerder = "fts_total_queries_rejected_by_herder"
)

type FTS struct {
	Op struct {
		Samples map[string][]float64 `json:"samples"`
	} `json:"op"`
}
