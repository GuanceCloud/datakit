/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowInstanceTagsResponse struct {
	// 标签列表
	Tags           *[]CreateInstanceReqTags `json:"tags,omitempty"`
	HttpStatusCode int                      `json:"-"`
}

func (o ShowInstanceTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceTagsResponse struct{}"
	}

	return strings.Join([]string{"ShowInstanceTagsResponse", string(data)}, " ")
}
