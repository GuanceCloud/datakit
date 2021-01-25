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
type BatchDeleteInstanceTopicResponse struct {
	// Topic列表。
	Topics         *[]BatchDeleteInstanceTopicRespTopics `json:"topics,omitempty"`
	HttpStatusCode int                                   `json:"-"`
}

func (o BatchDeleteInstanceTopicResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteInstanceTopicResponse struct{}"
	}

	return strings.Join([]string{"BatchDeleteInstanceTopicResponse", string(data)}, " ")
}
