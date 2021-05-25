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

// This is a auto create Body Object
type NovaDisassociateSecurityGroupRequestBody struct {
	RemoveSecurityGroup *NovaRemoveSecurityGroupOption `json:"removeSecurityGroup"`
}

func (o NovaDisassociateSecurityGroupRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaDisassociateSecurityGroupRequestBody struct{}"
	}

	return strings.Join([]string{"NovaDisassociateSecurityGroupRequestBody", string(data)}, " ")
}
