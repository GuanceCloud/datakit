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
type CreateScalingGroupResponse struct {
	// 伸缩组ID
	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateScalingGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingGroupResponse", string(data)}, " ")
}
