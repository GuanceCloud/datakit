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
type DeleteSecurityGroupRequest struct {
	SecurityGroupId string `json:"security_group_id"`
}

func (o DeleteSecurityGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"DeleteSecurityGroupRequest", string(data)}, " ")
}
