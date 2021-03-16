/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type RunEvaluationByPolicyAssignmentIdRequest struct {
	PolicyAssignmentId string `json:"policy_assignment_id"`
}

func (o RunEvaluationByPolicyAssignmentIdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RunEvaluationByPolicyAssignmentIdRequest struct{}"
	}

	return strings.Join([]string{"RunEvaluationByPolicyAssignmentIdRequest", string(data)}, " ")
}
