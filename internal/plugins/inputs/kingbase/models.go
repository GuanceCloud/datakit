// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

type connectionResult struct {
	ActiveConnections int64 `db:"active_connections"`
	IdleConnections   int64 `db:"idle_connections"`
	MaxConnections    int64 `db:"max_connections"`
}

type queryStat struct {
	QueryID        string  `db:"queryid"`
	TotalTime      float64 `db:"total_time"`
	Calls          int64   `db:"calls"`
	Rows           int64   `db:"rows"`
	SharedBlksHit  int64   `db:"shared_blks_hit"`
	SharedBlksRead int64   `db:"shared_blks_read"`
}

type bufferCache struct {
	SharedBlksHit  int64   `db:"shared_blks_hit"`
	SharedBlksRead int64   `db:"shared_blks_read"`
	BufferHitRatio float64 `db:"buffer_hit_ratio"`
}

type indexUsage struct {
	IdxScan       int64   `db:"idx_scan"`
	SeqScan       int64   `db:"seq_scan"`
	IndexHitRatio float64 `db:"index_hit_ratio"`
}

type backgroundWriter struct {
	BuffersClean     int64 `db:"buffers_clean"`
	BuffersBackend   int64 `db:"buffers_backend"`
	CheckpointsTimed int64 `db:"checkpoints_timed"`
	CheckpointsReq   int64 `db:"checkpoints_req"`
}

type transaction struct {
	Commits   int64 `db:"commits"`
	Rollbacks int64 `db:"rollbacks"`
}

type lock struct {
	WaitingLocks int64 `db:"waiting_locks"`
}

type tablespace struct {
	SpcName   string `db:"spcname"`
	SizeBytes int64  `db:"size_bytes"`
}

type queryPerformance struct {
	QueryID      string  `db:"queryid"`
	Query        string  `db:"query"`
	MeanExecTime float64 `db:"mean_exec_time"`
}

type databaseStatus struct {
	Datname     string `db:"datname"`
	Numbackends int64  `db:"numbackends"`
	BlksHit     int64  `db:"blks_hit"`
	BlksRead    int64  `db:"blks_read"`
	TupInserted int64  `db:"tup_inserted"`
	TupUpdated  int64  `db:"tup_updated"`
	TupDeleted  int64  `db:"tup_deleted"`
	Conflicts   int64  `db:"conflicts"`
}

type lockDetail struct {
	LockType  string `db:"lock_type"`
	LockCount int64  `db:"lock_count"`
}

type sessionActivity struct {
	State        string `db:"state"`
	WaitEvent    string `db:"wait_event"`
	SessionCount int64  `db:"session_count"`
}

type queryCancellation struct {
	TempFiles int64 `db:"temp_files"`
	Deadlocks int64 `db:"deadlocks"`
}

type FunctionStat struct {
	Schemaname string  `db:"schemaname"`
	Funcname   string  `db:"funcname"`
	Calls      int64   `db:"calls"`
	TotalTime  float64 `db:"total_time"`
	SelfTime   float64 `db:"self_time"`
}

type slowQuery struct {
	QueryID       string  `db:"queryid"`
	Query         string  `db:"query"`
	TotalExecTime float64 `db:"total_exec_time"`
	Calls         int64   `db:"calls"`
	MeanExecTime  float64 `db:"mean_exec_time"`
}
