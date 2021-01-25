/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ShowGroupsRespGroupGroupMessageOffsets struct {
	// 分区编号。
	Partition *int32 `json:"partition,omitempty"`
	// 剩余可消费消息数，即消息堆积数。
	Lag *int32 `json:"lag,omitempty"`
	// topic名称。
	Topic *string `json:"topic,omitempty"`
	// 当前消费进度。
	MessageCurrentOffset *int32 `json:"message_current_offset,omitempty"`
	// 最大消息位置（LEO）。
	MessageLogEndOffset *int32 `json:"message_log_end_offset,omitempty"`
}

func (o ShowGroupsRespGroupGroupMessageOffsets) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowGroupsRespGroupGroupMessageOffsets struct{}"
	}

	return strings.Join([]string{"ShowGroupsRespGroupGroupMessageOffsets", string(data)}, " ")
}
