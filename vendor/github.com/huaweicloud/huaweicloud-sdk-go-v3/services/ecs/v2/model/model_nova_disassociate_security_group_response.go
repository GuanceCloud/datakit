/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type NovaDisassociateSecurityGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NovaDisassociateSecurityGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaDisassociateSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"NovaDisassociateSecurityGroupResponse", string(data)}, " ")
}
