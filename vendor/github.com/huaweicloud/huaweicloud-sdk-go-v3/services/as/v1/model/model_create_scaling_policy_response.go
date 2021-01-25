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
type CreateScalingPolicyResponse struct {
	// 伸缩策略ID。
	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	HttpStatusCode  int     `json:"-"`
}

func (o CreateScalingPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateScalingPolicyResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingPolicyResponse", string(data)}, " ")
}
