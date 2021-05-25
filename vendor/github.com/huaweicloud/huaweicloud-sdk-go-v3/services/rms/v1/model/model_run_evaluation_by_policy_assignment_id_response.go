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

// Response Object
type RunEvaluationByPolicyAssignmentIdResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o RunEvaluationByPolicyAssignmentIdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RunEvaluationByPolicyAssignmentIdResponse struct{}"
	}

	return strings.Join([]string{"RunEvaluationByPolicyAssignmentIdResponse", string(data)}, " ")
}
