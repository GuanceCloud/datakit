/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListTopicsResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 返回的Topic个数。该参数不受offset和limit影响，即返回的是您账户下所有的Topic个数。
	TopicCount *int32 `json:"topic_count,omitempty"`
	// Topic结构体数组。
	Topics         *[]ListTopicsItem `json:"topics,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListTopicsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTopicsResponse struct{}"
	}

	return strings.Join([]string{"ListTopicsResponse", string(data)}, " ")
}
