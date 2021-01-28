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
type RevokeRequest struct {
	XLanguage  *string            `json:"X-Language,omitempty"`
	InstanceId string             `json:"instance_id"`
	Body       *RevokeRequestBody `json:"body,omitempty"`
}

func (o RevokeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RevokeRequest struct{}"
	}

	return strings.Join([]string{"RevokeRequest", string(data)}, " ")
}
