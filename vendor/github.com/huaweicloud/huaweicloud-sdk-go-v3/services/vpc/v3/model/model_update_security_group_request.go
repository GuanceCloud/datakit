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
type UpdateSecurityGroupRequest struct {
	SecurityGroupId string                          `json:"security_group_id"`
	Body            *UpdateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o UpdateSecurityGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"UpdateSecurityGroupRequest", string(data)}, " ")
}
