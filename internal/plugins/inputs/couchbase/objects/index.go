// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	// Samples Keys.
	IndexMemoryQuota  = "index_memory_quota"
	IndexMemoryUsed   = "index_memory_used"
	IndexRAMPercent   = "index_ram_percent"
	IndexRemainingRAM = "index_remaining_ram"

	// these are const keys for the Indexer Stats.
	IndexDocsIndexed          = "num_docs_indexed"
	IndexItemsCount           = "items_count"
	IndexFragPercent          = "frag_percent"
	IndexNumDocsPendingQueued = "num_docs_pending_queued"
	IndexNumRequests          = "num_requests"
	IndexCacheMisses          = "cache_misses"
	IndexCacheHits            = "cache_hits"
	IndexCacheHitPercent      = "cache_hit_percent"
	IndexNumRowsReturned      = "num_rows_returned"
	IndexResidentPercent      = "resident_percent"
	IndexAvgScanLatency       = "avg_scan_latency"
)

type Index struct {
	Op struct {
		Samples map[string][]float64 `json:"samples"`
	} `json:"op"`
}
