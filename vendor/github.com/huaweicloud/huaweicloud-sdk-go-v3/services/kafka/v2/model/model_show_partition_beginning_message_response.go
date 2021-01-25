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
type ShowPartitionBeginningMessageResponse struct {
	// Topic名称。
	Topic *string `json:"topic,omitempty"`
	// 分区编号。
	Partition *int32 `json:"partition,omitempty"`
	// 最新消息位置。
	MessageOffset  *int32 `json:"message_offset,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ShowPartitionBeginningMessageResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPartitionBeginningMessageResponse struct{}"
	}

	return strings.Join([]string{"ShowPartitionBeginningMessageResponse", string(data)}, " ")
}
