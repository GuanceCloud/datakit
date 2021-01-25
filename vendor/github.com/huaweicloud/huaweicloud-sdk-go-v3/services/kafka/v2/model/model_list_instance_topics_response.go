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
type ListInstanceTopicsResponse struct {
	// topic总数。
	Count *int32 `json:"count,omitempty"`
	// 分页查询的大小。
	Size *int32 `json:"size,omitempty"`
	// Topic列表。
	Topics         *[]ListInstanceTopicsRespTopics `json:"topics,omitempty"`
	HttpStatusCode int                             `json:"-"`
}

func (o ListInstanceTopicsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstanceTopicsResponse struct{}"
	}

	return strings.Join([]string{"ListInstanceTopicsResponse", string(data)}, " ")
}
