/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DescribeStreamRequest struct {
	StreamName       string  `json:"stream_name"`
	StartPartitionId *string `json:"start_partitionId,omitempty"`
	LimitPartitions  *int32  `json:"limit_partitions,omitempty"`
}

func (o DescribeStreamRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DescribeStreamRequest struct{}"
	}

	return strings.Join([]string{"DescribeStreamRequest", string(data)}, " ")
}
