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
type ShowPartitionMessageResponse struct {
	// 消息列表。
	Message        *[]ShowPartitionMessageRespMessage `json:"message,omitempty"`
	HttpStatusCode int                                `json:"-"`
}

func (o ShowPartitionMessageResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPartitionMessageResponse struct{}"
	}

	return strings.Join([]string{"ShowPartitionMessageResponse", string(data)}, " ")
}
