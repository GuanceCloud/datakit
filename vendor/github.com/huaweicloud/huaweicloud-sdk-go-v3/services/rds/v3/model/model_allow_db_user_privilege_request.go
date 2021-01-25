/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type AllowDbUserPrivilegeRequest struct {
	XLanguage  *string       `json:"X-Language,omitempty"`
	InstanceId string        `json:"instance_id"`
	Body       *GrantRequest `json:"body,omitempty"`
}

func (o AllowDbUserPrivilegeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AllowDbUserPrivilegeRequest struct{}"
	}

	return strings.Join([]string{"AllowDbUserPrivilegeRequest", string(data)}, " ")
}
