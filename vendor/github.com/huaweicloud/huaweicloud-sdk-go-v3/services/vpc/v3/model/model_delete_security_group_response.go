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
type DeleteSecurityGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteSecurityGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"DeleteSecurityGroupResponse", string(data)}, " ")
}
