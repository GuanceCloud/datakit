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

type PutRecordsResultEntry struct {
	// 数据上传到的分区ID。
	PartitionId *string `json:"partition_id,omitempty"`
	// 数据上传到的序列号。序列号是每个记录的唯一标识符。序列号由DIS在数据生产者调用PutRecords操作以添加数据到DIS数据通道时DIS服务自动分配的。同一分区键的序列号通常会随时间变化增加。PutRecords请求之间的时间段越长，序列号越大。
	SequenceNumber *string `json:"sequence_number,omitempty"`
	// 错误码。
	ErrorCode *string `json:"error_code,omitempty"`
	// 错误消息。
	ErrorMessage *string `json:"error_message,omitempty"`
}

func (o PutRecordsResultEntry) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PutRecordsResultEntry struct{}"
	}

	return strings.Join([]string{"PutRecordsResultEntry", string(data)}, " ")
}
