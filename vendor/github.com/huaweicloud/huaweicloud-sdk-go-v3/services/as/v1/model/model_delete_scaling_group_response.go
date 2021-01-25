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
type DeleteScalingGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingGroupResponse", string(data)}, " ")
}
