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
type UpdateScalingGroupResponse struct {
	// 伸缩组ID
	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateScalingGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"UpdateScalingGroupResponse", string(data)}, " ")
}
