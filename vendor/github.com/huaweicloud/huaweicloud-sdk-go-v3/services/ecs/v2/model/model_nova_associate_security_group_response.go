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
type NovaAssociateSecurityGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NovaAssociateSecurityGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaAssociateSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"NovaAssociateSecurityGroupResponse", string(data)}, " ")
}
