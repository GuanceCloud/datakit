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
type ListPolicyAssignmentsRequest struct {
}

func (o ListPolicyAssignmentsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPolicyAssignmentsRequest struct{}"
	}

	return strings.Join([]string{"ListPolicyAssignmentsRequest", string(data)}, " ")
}
