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
type UpdateScalingGroupInstanceResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateScalingGroupInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateScalingGroupInstanceResponse struct{}"
	}

	return strings.Join([]string{"UpdateScalingGroupInstanceResponse", string(data)}, " ")
}
