// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	// Keys for samples structures.
	CbasDiskUsed          = "cbas_disk_used"
	CbasGcCount           = "cbas_gc_count"
	CbasGcTime            = "cbas_gc_time"
	CbasHeapUsed          = "cbas_heap_used"
	CbasIoReads           = "cbas_io_reads"
	CbasIoWrites          = "cbas_io_writes"
	CbasSystemLoadAverage = "cbas_system_load_average"
	CbasThreadCount       = "cbas_thread_count"
)

type Analytics struct {
	Op struct {
		Samples      map[string][]float64 `json:"samples"`
		SamplesCount int                  `json:"samplesCount"`
		IsPersistent bool                 `json:"isPersistent"`
		LastTStamp   int64                `json:"lastTStamp"`
		Interval     int                  `json:"interval"`
	} `json:"op"`
}
