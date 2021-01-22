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
type CreateDbUserRequest struct {
	XLanguage  *string          `json:"X-Language,omitempty"`
	InstanceId string           `json:"instance_id"`
	Body       *UserForCreation `json:"body,omitempty"`
}

func (o CreateDbUserRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDbUserRequest struct{}"
	}

	return strings.Join([]string{"CreateDbUserRequest", string(data)}, " ")
}
