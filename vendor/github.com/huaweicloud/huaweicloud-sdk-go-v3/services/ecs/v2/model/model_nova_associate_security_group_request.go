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

// Request Object
type NovaAssociateSecurityGroupRequest struct {
	ServerId string                                 `json:"server_id"`
	Body     *NovaAssociateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o NovaAssociateSecurityGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaAssociateSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"NovaAssociateSecurityGroupRequest", string(data)}, " ")
}
