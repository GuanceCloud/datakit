// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	// Samples Keys.
	EventingCheckpointFailureCount     = "eventing/checkpoint_failure_count"
	EventingBucketOpExceptionCount     = "eventing/bucket_op_exception_count"
	EventingDcpBacklog                 = "eventing/dcp_backlog"
	EventingFailedCount                = "eventing/failed_count"
	EventingN1QlOpExceptionCount       = "eventing/n1ql_op_exception_count"
	EventingOnDeleteFailure            = "eventing/on_delete_failure"
	EventingOnDeleteSuccess            = "eventing/on_delete_success"
	EventingOnUpdateFailure            = "eventing/on_update_failure"
	EventingOnUpdateSuccess            = "eventing/on_update_success"
	EventingProcessedCount             = "eventing/processed_count"
	EventingTestBucketOpExceptionCount = "eventing/test/bucket_op_exception_count"
	EventingTestCheckpointFailureCount = "eventing/test/checkpoint_failure_count"
	EventingTestDcpBacklog             = "eventing/test/dcp_backlog"
	EventingTestFailedCount            = "eventing/test/failed_count"
	EventingTestN1QlOpExceptionCount   = "eventing/test/n1ql_op_exception_count"
	EventingTestOnDeleteFailure        = "eventing/test/on_delete_failure"
	EventingTestOnDeleteSuccess        = "eventing/test/on_delete_success"
	EventingTestOnUpdateFailure        = "eventing/test/on_update_failure"
	EventingTestOnUpdateSuccess        = "eventing/test/on_update_success"
	EventingTestProcessedCount         = "eventing/test/processed_count"
	EventingTestTimeoutCount           = "eventing/test/timeout_count"
	EventingTimeoutCount               = "eventing/timeout_count"
)

type Eventing struct {
	Op struct {
		Samples map[string][]float64 `json:"samples"`
	} `json:"op"`
}
