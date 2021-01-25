/*
 * DDS
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
type CreateDatabaseRoleRequest struct {
	InstanceId string                         `json:"instance_id"`
	Body       *CreateDatabaseRoleRequestBody `json:"body,omitempty"`
}

func (o CreateDatabaseRoleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatabaseRoleRequest struct{}"
	}

	return strings.Join([]string{"CreateDatabaseRoleRequest", string(data)}, " ")
}
