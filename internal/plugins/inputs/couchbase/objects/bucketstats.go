// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	// Samples Keys
	// The size of all data files for this bucket, including the data
	// itself, meta data and temporary files.
	CouchTotalDiskSize = "couch_total_disk_size"
	// How much fragmented data there is to be compacted compared
	// to real data for the data files in this bucket.
	CouchDocsFragmentation = "couch_docs_fragmentation"
	// How much fragmented data there is to be compacted compared
	// to real data for the view index files in this bucket.
	CouchViewsFragmentation = "couch_views_fragmentation"

	HitRatio = "hit_ratio"
	// Percentage of reads per second to this bucket
	// from disk as opposed to RAM (measured from
	// ep_bg_fetches / cmd_total_gets * 100).
	EpCacheMissRate = "ep_cache_miss_rate"
	// Percentage of all items cached in RAM in this bucket.
	EpResidentItemsRate                 = "ep_resident_items_rate"
	VbAvgActiveQueueAge                 = "vb_avg_active_queue_age"
	VbAvgReplicaQueueAge                = "vb_avg_replica_queue_age"
	VbAvgPendingQueueAge                = "vb_avg_pending_queue_age"
	VbAvgTotalQueueAge                  = "vb_avg_total_queue_age"
	VbActiveResidentItemsRatio          = "vb_active_resident_items_ratio"
	VbReplicaResidentItemsRatio         = "vb_replica_resident_items_ratio"
	VbPendingResidentItemsRatio         = "vb_pending_resident_items_ratio"
	AvgDiskUpdateTime                   = "avg_disk_update_time"
	AvgDiskCommitTime                   = "avg_disk_commit_time"
	AvgBgWaitTime                       = "avg_bg_wait_time"
	AvgActiveTimestampDrift             = "avg_active_timestamp_drift"  // couchbase 5.1.1
	AvgReplicaTimestampDrift            = "avg_replica_timestamp_drift" // couchbase 5.1.1
	EpDcpViewsIndexesCount              = "ep_dcp_views+indexes_count"
	EpDcpViewsIndexesItemsRemaining     = "ep_dcp_views+indexes_items_remaining"
	EpDcpViewsIndexesProducerCount      = "ep_dcp_views+indexes_producer_count"
	EpDcpViewsIndexesTotalBacklogSize   = "ep_dcp_views+indexes_total_backlog_size"
	EpDcpViewsIndexesItemsSent          = "ep_dcp_views+indexes_items_sent"
	EpDcpViewsIndexesTotalBytes         = "ep_dcp_views+indexes_total_bytes"
	EpDcpViewsIndexesBackoff            = "ep_dcp_views+indexes_backoff"
	BgWaitCount                         = "bg_wait_count"
	BgWaitTotal                         = "bg_wait_total"
	BytesRead                           = "bytes_read"
	BytesWritten                        = "bytes_written"
	CasBadval                           = "cas_badval"
	CasHits                             = "cas_hits"
	CasMisses                           = "cas_misses"
	BucketStatsCmdGet                   = "cmd_get"
	CmdSet                              = "cmd_set"
	BucketStatsCouchDocsActualDiskSize  = "couch_docs_actual_disk_size"
	BucketStatsCouchDocsDataSize        = "couch_docs_data_size"
	CouchDocsDiskSize                   = "couch_docs_disk_size"
	BucketStatsCouchSpatialDataSize     = "couch_spatial_data_size"
	BucketStatsCouchSpatialDiskSize     = "couch_spatial_disk_size"
	CouchSpatialOps                     = "couch_spatial_ops"
	BucketStatsCouchViewsActualDiskSize = "couch_views_actual_disk_size"
	BucketStatsCouchViewsDataSize       = "couch_views_data_size"
	CouchViewsDiskSize                  = "couch_views_disk_size"
	CouchViewsOps                       = "couch_views_ops"
	CurrConnections                     = "curr_connections"
	BucketStatsCurrItems                = "curr_items"
	BucketStatsCurrItemsTot             = "curr_items_tot"
	DecrHits                            = "decr_hits"
	DecrMisses                          = "decr_misses"
	DeleteHits                          = "delete_hits"
	DeleteMisses                        = "delete_misses"
	DiskCommitCount                     = "disk_commit_count"
	DiskCommitTotal                     = "disk_commit_total"
	DiskUpdateCount                     = "disk_update_count"
	DiskUpdateTotal                     = "disk_update_total"
	DiskWriteQueue                      = "disk_write_queue"
	EpActiveAheadExceptions             = "ep_active_ahead_exceptions"
	EpActiveHlcDrift                    = "ep_active_hlc_drift"
	EpActiveHlcDriftCount               = "ep_active_hlc_drift_count"
	BucketStatsEpBgFetched              = "ep_bg_fetched"
	EpClockCasDriftThresholdExceeded    = "ep_clock_cas_drift_threshold_exceeded"
	EpDcp2IBackoff                      = "ep_dcp_2i_backoff"
	EpDcp2ICount                        = "ep_dcp_2i_count"
	EpDcp2IItemsRemaining               = "ep_dcp_2i_items_remaining"
	EpDcp2IItemsSent                    = "ep_dcp_2i_items_sent"
	EpDcp2IProducerCount                = "ep_dcp_2i_producer_count"
	EpDcp2ITotalBacklogSize             = "ep_dcp_2i_total_backlog_size"
	EpDcp2ITotalBytes                   = "ep_dcp_2i_total_bytes"
	EpDcpFtsBackoff                     = "ep_dcp_fts_backoff"
	EpDcpFtsCount                       = "ep_dcp_fts_count"
	EpDcpFtsItemsRemaining              = "ep_dcp_fts_items_remaining"
	EpDcpFtsItemsSent                   = "ep_dcp_fts_items_sent"
	EpDcpFtsProducerCount               = "ep_dcp_fts_producer_count"
	EpDcpFtsTotalBacklogSize            = "ep_dcp_fts_total_backlog_size"
	EpDcpFtsTotalBytes                  = "ep_dcp_fts_total_bytes"
	EpDcpOtherBackoff                   = "ep_dcp_other_backoff"
	EpDcpOtherCount                     = "ep_dcp_other_count"
	EpDcpOtherItemsRemaining            = "ep_dcp_other_items_remaining"
	EpDcpOtherItemsSent                 = "ep_dcp_other_items_sent"
	EpDcpOtherProducerCount             = "ep_dcp_other_producer_count"
	EpDcpOtherTotalBacklogSize          = "ep_dcp_other_total_backlog_size"
	EpDcpOtherTotalBytes                = "ep_dcp_other_total_bytes"
	EpDcpReplicaBackoff                 = "ep_dcp_replica_backoff"
	EpDcpReplicaCount                   = "ep_dcp_replica_count"
	EpDcpReplicaItemsRemaining          = "ep_dcp_replica_items_remaining"
	EpDcpReplicaItemsSent               = "ep_dcp_replica_items_sent"
	EpDcpReplicaProducerCount           = "ep_dcp_replica_producer_count"
	EpDcpReplicaTotalBacklogSize        = "ep_dcp_replica_total_backlog_size"
	EpDcpReplicaTotalBytes              = "ep_dcp_replica_total_bytes"
	EpDcpViewsBackoff                   = "ep_dcp_views_backoff"
	EpDcpViewsCount                     = "ep_dcp_views_count"
	EpDcpViewsItemsRemaining            = "ep_dcp_views_items_remaining"
	EpDcpViewsItemsSent                 = "ep_dcp_views_items_sent"
	EpDcpViewsProducerCount             = "ep_dcp_views_producer_count"
	EpDcpViewsTotalBacklogSize          = "ep_dcp_views_total_backlog_size"
	EpDcpViewsTotalBytes                = "ep_dcp_views_total_bytes"
	EpDcpXdcrBackoff                    = "ep_dcp_xdcr_backoff"
	EpDcpXdcrCount                      = "ep_dcp_xdcr_count"
	EpDcpXdcrItemsRemaining             = "ep_dcp_xdcr_items_remaining"
	EpDcpXdcrItemsSent                  = "ep_dcp_xdcr_items_sent"
	EpDcpXdcrProducerCount              = "ep_dcp_xdcr_producer_count"
	EpDcpXdcrTotalBacklogSize           = "ep_dcp_xdcr_total_backlog_size"
	EpDcpXdcrTotalBytes                 = "ep_dcp_xdcr_total_bytes"
	EpDiskqueueDrain                    = "ep_diskqueue_drain"
	EpDiskqueueFill                     = "ep_diskqueue_fill"
	EpDiskqueueItems                    = "ep_diskqueue_items"
	EpFlusherTodo                       = "ep_flusher_todo"
	EpItemCommitFailed                  = "ep_item_commit_failed"
	EpKvSize                            = "ep_kv_size"
	EpMaxSize                           = "ep_max_size"
	EpMemHighWat                        = "ep_mem_high_wat"
	EpMemLowWat                         = "ep_mem_low_wat"
	EpMetaDataMemory                    = "ep_meta_data_memory"
	EpNumNonResident                    = "ep_num_non_resident"
	EpNumOpsDelMeta                     = "ep_num_ops_del_meta"
	EpNumOpsDelRetMeta                  = "ep_num_ops_del_ret_meta"
	EpNumOpsGetMeta                     = "ep_num_ops_get_meta"
	EpNumOpsSetMeta                     = "ep_num_ops_set_meta"
	EpNumOpsSetRetMeta                  = "ep_num_ops_set_ret_meta"
	EpNumValueEjects                    = "ep_num_value_ejects"
	EpOomErrors                         = "ep_oom_errors"
	EpOpsCreate                         = "ep_ops_create"
	EpOpsUpdate                         = "ep_ops_update"
	EpOverhead                          = "ep_overhead"
	EpQueueSize                         = "ep_queue_size"
	EpReplicaAheadExceptions            = "ep_replica_ahead_exceptions"
	EpReplicaHlcDrift                   = "ep_replica_hlc_drift"
	EpReplicaHlcDriftCount              = "ep_replica_hlc_drift_count"
	EpTmpOomErrors                      = "ep_tmp_oom_errors"
	EpVbTotal                           = "ep_vb_total"
	Evictions                           = "evictions"
	BucketStatsGetHits                  = "get_hits"
	GetMisses                           = "get_misses"
	IncrHits                            = "incr_hits"
	IncrMisses                          = "incr_misses"
	BucketStatsMemUsed                  = "mem_used"
	Misses                              = "misses"
	BucketStatsOps                      = "ops"
	XdcOps                              = "xdc_ops"
	VbActiveEject                       = "vb_active_eject"
	VbActiveItmMemory                   = "vb_active_itm_memory"
	VbActiveMetaDataMemory              = "vb_active_meta_data_memory"
	VbActiveNum                         = "vb_active_num"
	BucketStatsVbActiveNumNonResident   = "vb_active_num_non_resident"
	VbActiveOpsCreate                   = "vb_active_ops_create"
	VbActiveOpsUpdate                   = "vb_active_ops_update"
	VbActiveQueueAge                    = "vb_active_queue_age"
	VbActiveQueueDrain                  = "vb_active_queue_drain"
	VbActiveQueueFill                   = "vb_active_queue_fill"
	VbActiveQueueSize                   = "vb_active_queue_size"
	VbPendingCurrItems                  = "vb_pending_curr_items"
	VbPendingEject                      = "vb_pending_eject"
	VbPendingItmMemory                  = "vb_pending_itm_memory"
	VbPendingMetaDataMemory             = "vb_pending_meta_data_memory"
	VbPendingNum                        = "vb_pending_num"
	VbPendingNumNonResident             = "vb_pending_num_non_resident"
	VbPendingOpsCreate                  = "vb_pending_ops_create"
	VbPendingOpsUpdate                  = "vb_pending_ops_update"
	VbPendingQueueAge                   = "vb_pending_queue_age"
	VbPendingQueueDrain                 = "vb_pending_queue_drain"
	VbPendingQueueFill                  = "vb_pending_queue_fill"
	VbPendingQueueSize                  = "vb_pending_queue_size"
	BucketStatsVbReplicaCurrItems       = "vb_replica_curr_items"
	VbReplicaEject                      = "vb_replica_eject"
	VbReplicaItmMemory                  = "vb_replica_itm_memory"
	VbReplicaMetaDataMemory             = "vb_replica_meta_data_memory"
	VbReplicaNum                        = "vb_replica_num"
	VbReplicaNumNonResident             = "vb_replica_num_non_resident"
	VbReplicaOpsCreate                  = "vb_replica_ops_create"
	VbReplicaOpsUpdate                  = "vb_replica_ops_update"
	VbReplicaQueueAge                   = "vb_replica_queue_age"
	VbReplicaQueueDrain                 = "vb_replica_queue_drain"
	VbReplicaQueueFill                  = "vb_replica_queue_fill"
	VbReplicaQueueSize                  = "vb_replica_queue_size"
	VbTotalQueueAge                     = "vb_total_queue_age"
	CPUIdleMs                           = "cpu_idle_ms"
	CPULocalMs                          = "cpu_local_ms"
	BucketStatsCPUUtilizationRate       = "cpu_utilization_rate"
	HibernatedRequests                  = "hibernated_requests"
	HibernatedWaked                     = "hibernated_waked"
	MemActualFree                       = "mem_actual_free"
	MemActualUsed                       = "mem_actual_used"
	BucketStatsMemFree                  = "mem_free"
	BucketStatsMemTotal                 = "mem_total"
	MemUsedSys                          = "mem_used_sys"
	RestRequests                        = "rest_requests"
	BucketStatsSwapTotal                = "swap_total"
	BucketStatsSwapUsed                 = "swap_used"

	// these keys are not present in 6.6.2 and I believe they have been deprecated.
	DEPRECATEDEpDcpCbasBackoff          = "ep_dcp_cbas_backoff"
	DEPRECATEDEpDcpCbasItemsRemaining   = "ep_dcp_cbas_items_remaining"
	DEPRECATEDEpDcpTotalBytes           = "ep_dcp_total_bytes"
	DEPRECATEDEpDcpCbasTotalBacklogSize = "ep_dcp_cbas_total_backlog_size"
	DEPRECATEDEpDataWriteFailed         = "ep_data_write_failed"
	DEPRECATEDEpDataReadFailed          = "ep_data_read_failed"
	DEPRECATEDEpDcpCbasProducerCount    = "ep_dcp_cbas_producer_count"
	DEPRECATEDEpDcpCbasCount            = "ep_dcp_cbas_count"
	DEPRECATEDEpDcpCbasItemsSent        = "ep_dcp_cbas_items_sent"
	DEPRECATEDVbActiveQuueItems         = "vb_active_queue_items"
)

// PerNodeBucketStats separate struct as the Samples needs to be a map[string]interface{}.
// /pools/default/buckets/<bucket-name>/nodes/<node-name>/stats.
type PerNodeBucketStats struct {
	HostName string `json:"hostname,omitempty"` // per node stats only
	Op       struct {
		Samples      map[string]interface{} `json:"samples"`
		SamplesCount int                    `json:"samplesCount"`
		IsPersistent bool                   `json:"isPersistent"`
		LastTStamp   int64                  `json:"lastTStamp"`
		Interval     int                    `json:"interval"`
	} `json:"op"`
	HotKeys []struct {
		Name string  `json:"name"`
		Ops  float64 `json:"ops"`
	} `json:"hot_keys,omitempty"`
}

// /pools/default/buckets/<bucket-name>/stats only
// for /pools/default/buckets/<bucket-name>/nodes/<node-name>/stats see PerNodeBucketStats

type BucketStats struct {
	Op struct {
		Samples      map[string][]float64 `json:"samples"`
		SamplesCount float64              `json:"samplesCount"`
		IsPersistent bool                 `json:"isPersistent"`
		LastTStamp   float64              `json:"lastTStamp"`
		Interval     float64              `json:"interval"`
	} `json:"op"`
	HotKeys []struct {
		Name string  `json:"name"`
		Ops  float64 `json:"ops"`
	} `json:"hot_keys,omitempty"`
}
