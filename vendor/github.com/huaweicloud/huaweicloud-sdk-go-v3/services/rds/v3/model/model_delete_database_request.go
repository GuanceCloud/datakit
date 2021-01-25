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
type DeleteDatabaseRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	DbName     string  `json:"db_name"`
}

func (o DeleteDatabaseRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteDatabaseRequest struct{}"
	}

	return strings.Join([]string{"DeleteDatabaseRequest", string(data)}, " ")
}
