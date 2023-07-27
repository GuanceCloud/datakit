// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

// Task task struct.
// /pools/default/tasks.
type Task struct {
	StatusID      string  `json:"statusId"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	RefreshPeriod float64 `json:"recommendedRefreshPeriod,omitempty"`
	Progress      float64 `json:"progress,omitempty"`

	// rebalance running
	SubType string `json:"subtype,omitempty"`
	PerNode map[string]struct {
		Progress float64 `json:"progress,omitempty"`
	} `json:"perNode,omitempty"`
	DetailedProgress struct {
		Bucket       string `json:"bucket,omitempty"`
		BucketNumber int    `json:"bucketNumber,omitempty"`
		BucketCount  int    `json:"bucketCount,omitempty"`
		PerNode      map[string]struct {
			Ingoing  NodeProgress `json:"ingoing,omitempty"`
			Outgoing NodeProgress `json:"outgoing,omitempty"`
		} `json:"perNode,omitempty"`
	} `json:"detailedProgress,omitempty"`
	StageInfo struct {
		Analytics StageInfo `json:"analytics,omitempty"`
		Eventing  StageInfo `json:"eventing,omitempty"`
		Search    StageInfo `json:"search,omitempty"`
		Index     StageInfo `json:"index,omitempty"`
		Data      StageInfo `json:"data,omitempty"`
		Query     StageInfo `json:"query,omitempty"`
	} `json:"stageInfo,omitempty"`

	// rebalance not running
	StatusIsStale         bool `json:"statusIsStale"`
	MasterRequestTimedOut bool `json:"masterRequestTimedOut"`

	// compacting stuff
	Bucket       string `json:"bucket,omitempty"`
	ChangesDone  int64  `json:"changesDone,omitempty"`
	TotalChanges int64  `json:"totalChanges,omitempty"`

	// XDCR stuff
	ChangesLeft    int64         `json:"changesLeft,omitempty"`
	DocsChecked    int64         `json:"docsChecked,omitempty"`
	DocsWritten    int64         `json:"docsWritten,omitempty"`
	PauseRequested bool          `json:"pauseRequested,omitempty"`
	Continuous     bool          `json:"continuous,omitempty"`
	Source         string        `json:"source,omitempty"`
	Target         string        `json:"target,omitempty"`
	Errors         []interface{} `json:"errors,omitempty"`
	MaxVBReps      string

	// loadingSampleBucket
	Pid string `json:"pid,omitempty"`
}

// NodeProgress is the ingoing/outgoing detailed progress of a task in a node.
type NodeProgress struct {
	DocsTotal           int64 `json:"docsTotal,omitempty"`
	DocsTransferred     int64 `json:"docsTransferred,omitempty"`
	ActiveVBucketsLeft  int64 `json:"activeVBucketsLeft,omitempty"`
	ReplicaVBucketsLeft int64 `json:"replicaVBucketsLeft,omitempty"`
}

type StageInfo struct {
	TotalProgress   int64         `json:"totalProgress,omitempty"`
	PerNodeProgress []interface{} `json:"perNodeProgress,omitempty"`
	StartTime       string        `json:"startTime,omitempty"`
	CompletedTime   string        `json:"completedTime,omitempty"`
	TimeTaken       int64         `json:"timeTaken,omitempty"`
}
