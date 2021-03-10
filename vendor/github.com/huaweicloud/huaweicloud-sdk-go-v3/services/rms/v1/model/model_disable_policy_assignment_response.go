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
type DisablePolicyAssignmentResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DisablePolicyAssignmentResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisablePolicyAssignmentResponse struct{}"
	}

	return strings.Join([]string{"DisablePolicyAssignmentResponse", string(data)}, " ")
}
