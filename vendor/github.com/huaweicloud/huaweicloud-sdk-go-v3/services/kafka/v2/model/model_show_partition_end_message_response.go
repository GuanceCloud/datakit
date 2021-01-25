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

// Response Object
type ShowPartitionEndMessageResponse struct {
	// Topic名称。
	Topic *string `json:"topic,omitempty"`
	// 分区编号。
	Partition *int32 `json:"partition,omitempty"`
	// 最新消息位置。
	MessageOffset  *int32 `json:"message_offset,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ShowPartitionEndMessageResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPartitionEndMessageResponse struct{}"
	}

	return strings.Join([]string{"ShowPartitionEndMessageResponse", string(data)}, " ")
}
