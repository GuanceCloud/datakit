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
type EnablePolicyAssignmentResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o EnablePolicyAssignmentResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnablePolicyAssignmentResponse struct{}"
	}

	return strings.Join([]string{"EnablePolicyAssignmentResponse", string(data)}, " ")
}
