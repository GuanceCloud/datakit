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
type CreatePostPaidInstanceResponse struct {
	// 实例ID。
	InstanceId     *string `json:"instance_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreatePostPaidInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePostPaidInstanceResponse struct{}"
	}

	return strings.Join([]string{"CreatePostPaidInstanceResponse", string(data)}, " ")
}
