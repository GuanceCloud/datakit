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

// Response Object
type ShowSecurityGroupResponse struct {
	SecurityGroup  *SecurityGroup `json:"security_group,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowSecurityGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"ShowSecurityGroupResponse", string(data)}, " ")
}
