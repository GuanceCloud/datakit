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
type CompleteLifecycleActionRequest struct {
	ScalingGroupId string                              `json:"scaling_group_id"`
	Body           *CompleteLifecycleActionRequestBody `json:"body,omitempty"`
}

func (o CompleteLifecycleActionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CompleteLifecycleActionRequest struct{}"
	}

	return strings.Join([]string{"CompleteLifecycleActionRequest", string(data)}, " ")
}
