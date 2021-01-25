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

type BatchDeleteInstanceTopicRespTopics struct {
	// Topic名称。
	Id *string `json:"id,omitempty"`
	// topic名称。
	Success *bool `json:"success,omitempty"`
}

func (o BatchDeleteInstanceTopicRespTopics) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteInstanceTopicRespTopics struct{}"
	}

	return strings.Join([]string{"BatchDeleteInstanceTopicRespTopics", string(data)}, " ")
}
