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
type ListAuthorizedDatabasesRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	UserName   string  `json:"user-name"`
	Page       int32   `json:"page"`
	Limit      int32   `json:"limit"`
}

func (o ListAuthorizedDatabasesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAuthorizedDatabasesRequest struct{}"
	}

	return strings.Join([]string{"ListAuthorizedDatabasesRequest", string(data)}, " ")
}
