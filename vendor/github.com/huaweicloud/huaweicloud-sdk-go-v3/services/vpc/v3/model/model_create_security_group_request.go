/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateSecurityGroupRequest struct {
	Body *CreateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o CreateSecurityGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"CreateSecurityGroupRequest", string(data)}, " ")
}
