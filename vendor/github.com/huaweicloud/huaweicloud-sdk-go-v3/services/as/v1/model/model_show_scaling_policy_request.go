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
type ShowScalingPolicyRequest struct {
	ScalingPolicyId string `json:"scaling_policy_id"`
}

func (o ShowScalingPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowScalingPolicyRequest struct{}"
	}

	return strings.Join([]string{"ShowScalingPolicyRequest", string(data)}, " ")
}
