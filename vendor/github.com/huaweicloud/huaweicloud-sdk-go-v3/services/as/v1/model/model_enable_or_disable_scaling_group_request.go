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
type EnableOrDisableScalingGroupRequest struct {
	ScalingGroupId string                                  `json:"scaling_group_id"`
	Body           *EnableOrDisableScalingGroupRequestBody `json:"body,omitempty"`
}

func (o EnableOrDisableScalingGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnableOrDisableScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"EnableOrDisableScalingGroupRequest", string(data)}, " ")
}
