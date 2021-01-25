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

// 消息体。
type ShowPartitionMessageRespMessage struct {
	// 消息的key。
	Key *string `json:"key,omitempty"`
	// 消息内容。
	Value *string `json:"value,omitempty"`
	// Topic名称。
	Topic *string `json:"topic,omitempty"`
	// 分区编号。
	Partition *int32 `json:"partition,omitempty"`
	// 消息位置。
	MessageOffset *int32 `json:"message_offset,omitempty"`
	// 消息大小，单位字节。
	Size *int32 `json:"size,omitempty"`
}

func (o ShowPartitionMessageRespMessage) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPartitionMessageRespMessage struct{}"
	}

	return strings.Join([]string{"ShowPartitionMessageRespMessage", string(data)}, " ")
}
