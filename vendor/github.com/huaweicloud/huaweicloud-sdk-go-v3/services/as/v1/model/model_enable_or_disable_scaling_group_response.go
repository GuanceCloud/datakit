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
type EnableOrDisableScalingGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o EnableOrDisableScalingGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnableOrDisableScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"EnableOrDisableScalingGroupResponse", string(data)}, " ")
}
