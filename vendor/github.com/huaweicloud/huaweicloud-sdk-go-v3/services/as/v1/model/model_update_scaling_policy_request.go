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
type UpdateScalingPolicyRequest struct {
	ScalingPolicyId string                          `json:"scaling_policy_id"`
	Body            *UpdateScalingPolicyRequestBody `json:"body,omitempty"`
}

func (o UpdateScalingPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateScalingPolicyRequest struct{}"
	}

	return strings.Join([]string{"UpdateScalingPolicyRequest", string(data)}, " ")
}
