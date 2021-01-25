/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type BatchDeleteScalingConfigsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteScalingConfigsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteScalingConfigsResponse struct{}"
	}

	return strings.Join([]string{"BatchDeleteScalingConfigsResponse", string(data)}, " ")
}
