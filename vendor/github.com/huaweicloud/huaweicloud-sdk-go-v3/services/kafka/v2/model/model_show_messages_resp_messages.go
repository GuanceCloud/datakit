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

type ShowMessagesRespMessages struct {
	// topic名称。
	Topic *string `json:"topic,omitempty"`
	// 分区编号。
	Partition *int32 `json:"partition,omitempty"`
	// 消息编号。
	MessageOffset *int32 `json:"message_offset,omitempty"`
	// 消息大小，单位字节。
	Size *int32 `json:"size,omitempty"`
}

func (o ShowMessagesRespMessages) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowMessagesRespMessages struct{}"
	}

	return strings.Join([]string{"ShowMessagesRespMessages", string(data)}, " ")
}
