/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateQueueResponse struct {
	// 队列ID。
	Id *string `json:"id,omitempty"`
	// 队列的名称。
	Name *string `json:"name,omitempty"`
	// 仅Kafka队列才有该响应参数。  使用Kafka SDK时的Kafka topic的ID。
	KafkaTopic     *string `json:"kafka_topic,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateQueueResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateQueueResponse struct{}"
	}

	return strings.Join([]string{"CreateQueueResponse", string(data)}, " ")
}
