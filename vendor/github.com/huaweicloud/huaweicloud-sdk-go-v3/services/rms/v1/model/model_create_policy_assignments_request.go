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
type CreatePolicyAssignmentsRequest struct {
	Body *PolicyAssignmentRequestBody `json:"body,omitempty"`
}

func (o CreatePolicyAssignmentsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePolicyAssignmentsRequest struct{}"
	}

	return strings.Join([]string{"CreatePolicyAssignmentsRequest", string(data)}, " ")
}
