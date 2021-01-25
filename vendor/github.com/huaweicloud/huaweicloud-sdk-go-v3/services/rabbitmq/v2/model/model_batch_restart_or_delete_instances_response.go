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
type BatchRestartOrDeleteInstancesResponse struct {
	// 修改实例的结果。
	Results        *[]BatchRestartOrDeleteInstanceRespResults `json:"results,omitempty"`
	HttpStatusCode int                                        `json:"-"`
}

func (o BatchRestartOrDeleteInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchRestartOrDeleteInstancesResponse struct{}"
	}

	return strings.Join([]string{"BatchRestartOrDeleteInstancesResponse", string(data)}, " ")
}
