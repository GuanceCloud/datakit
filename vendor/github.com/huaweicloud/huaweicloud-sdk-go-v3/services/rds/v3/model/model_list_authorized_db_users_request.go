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
type ListAuthorizedDbUsersRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	DbName     string  `json:"db-name"`
	Page       int32   `json:"page"`
	Limit      int32   `json:"limit"`
}

func (o ListAuthorizedDbUsersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAuthorizedDbUsersRequest struct{}"
	}

	return strings.Join([]string{"ListAuthorizedDbUsersRequest", string(data)}, " ")
}
