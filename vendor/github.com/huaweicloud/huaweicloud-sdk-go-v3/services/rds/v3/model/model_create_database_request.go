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
type CreateDatabaseRequest struct {
	XLanguage  *string              `json:"X-Language,omitempty"`
	InstanceId string               `json:"instance_id"`
	Body       *DatabaseForCreation `json:"body,omitempty"`
}

func (o CreateDatabaseRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatabaseRequest struct{}"
	}

	return strings.Join([]string{"CreateDatabaseRequest", string(data)}, " ")
}
