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
type ListTopicAttributesResponse struct {
	// 请求的唯一标识ID。
	RequestId      *string         `json:"request_id,omitempty"`
	Attributes     *TopicAttribute `json:"attributes,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListTopicAttributesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTopicAttributesResponse struct{}"
	}

	return strings.Join([]string{"ListTopicAttributesResponse", string(data)}, " ")
}
