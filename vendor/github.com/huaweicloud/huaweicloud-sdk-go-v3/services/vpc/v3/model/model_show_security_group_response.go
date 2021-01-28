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
	// 请求ID
	RequestId      *string            `json:"request_id,omitempty"`
	SecurityGroup  *SecurityGroupInfo `json:"security_group,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ShowSecurityGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"ShowSecurityGroupResponse", string(data)}, " ")
}
