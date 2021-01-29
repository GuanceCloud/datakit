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
type DeleteDbUserRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	UserName   string  `json:"user_name"`
}

func (o DeleteDbUserRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteDbUserRequest struct{}"
	}

	return strings.Join([]string{"DeleteDbUserRequest", string(data)}, " ")
}
