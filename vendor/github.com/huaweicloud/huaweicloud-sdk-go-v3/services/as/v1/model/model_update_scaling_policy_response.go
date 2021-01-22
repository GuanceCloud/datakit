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
type UpdateScalingPolicyResponse struct {
	// 伸缩策略ID。
	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	HttpStatusCode  int     `json:"-"`
}

func (o UpdateScalingPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateScalingPolicyResponse struct{}"
	}

	return strings.Join([]string{"UpdateScalingPolicyResponse", string(data)}, " ")
}
