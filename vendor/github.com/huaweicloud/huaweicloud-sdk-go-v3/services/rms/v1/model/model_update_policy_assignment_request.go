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
type UpdatePolicyAssignmentRequest struct {
	PolicyAssignmentId string                       `json:"policy_assignment_id"`
	Body               *PolicyAssignmentRequestBody `json:"body,omitempty"`
}

func (o UpdatePolicyAssignmentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePolicyAssignmentRequest struct{}"
	}

	return strings.Join([]string{"UpdatePolicyAssignmentRequest", string(data)}, " ")
}
