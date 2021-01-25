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
type ShowScalingGroupResponse struct {
	ScalingGroup   *ScalingGroups `json:"scaling_group,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowScalingGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowScalingGroupResponse struct{}"
	}

	return strings.Join([]string{"ShowScalingGroupResponse", string(data)}, " ")
}
