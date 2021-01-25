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

// Request Object
type UpdateScalingGroupRequest struct {
	ScalingGroupId string                         `json:"scaling_group_id"`
	Body           *UpdateScalingGroupRequestBody `json:"body,omitempty"`
}

func (o UpdateScalingGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"UpdateScalingGroupRequest", string(data)}, " ")
}
