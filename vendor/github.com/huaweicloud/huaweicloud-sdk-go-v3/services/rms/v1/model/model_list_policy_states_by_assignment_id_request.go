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
type ListPolicyStatesByAssignmentIdRequest struct {
	PolicyAssignmentId string  `json:"policy_assignment_id"`
	ComplianceState    *string `json:"compliance_state,omitempty"`
	ResourceId         *string `json:"resource_id,omitempty"`
	ResourceName       *string `json:"resource_name,omitempty"`
	Limit              *int32  `json:"limit,omitempty"`
	Marker             *string `json:"marker,omitempty"`
}

func (o ListPolicyStatesByAssignmentIdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPolicyStatesByAssignmentIdRequest struct{}"
	}

	return strings.Join([]string{"ListPolicyStatesByAssignmentIdRequest", string(data)}, " ")
}
