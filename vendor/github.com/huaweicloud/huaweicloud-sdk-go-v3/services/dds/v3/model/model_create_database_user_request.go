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
type CreateDatabaseUserRequest struct {
	InstanceId string                         `json:"instance_id"`
	Body       *CreateDatabaseUserRequestBody `json:"body,omitempty"`
}

func (o CreateDatabaseUserRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatabaseUserRequest struct{}"
	}

	return strings.Join([]string{"CreateDatabaseUserRequest", string(data)}, " ")
}
