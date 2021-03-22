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
type DeletePolicyAssignmentResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeletePolicyAssignmentResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePolicyAssignmentResponse struct{}"
	}

	return strings.Join([]string{"DeletePolicyAssignmentResponse", string(data)}, " ")
}
