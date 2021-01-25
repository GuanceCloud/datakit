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

type UpdatePartitionCount struct {
	// 扩缩容操作执行时间戳，13位时间戳。
	CreateTimestamp *int64 `json:"create_timestamp,omitempty"`
	// 扩缩容操作前分区数量。
	SrcPartitionCount *int32 `json:"src_partition_count,omitempty"`
	// 扩缩容操作后分区数量。
	TargetPartitionCount *int32 `json:"target_partition_count,omitempty"`
	// 扩缩容操作响应码。
	ResultCode *int32 `json:"result_code,omitempty"`
	// 扩缩容操作响应信息。
	ResultMsg *int32 `json:"result_msg,omitempty"`
	// 本次扩缩容操作是否为自动扩缩容。  - true：自动扩缩容。 - false：手动扩缩容。
	AutoScale *bool `json:"auto_scale,omitempty"`
}

func (o UpdatePartitionCount) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePartitionCount struct{}"
	}

	return strings.Join([]string{"UpdatePartitionCount", string(data)}, " ")
}
