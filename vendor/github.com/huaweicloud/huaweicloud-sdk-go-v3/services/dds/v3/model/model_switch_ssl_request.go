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
type SwitchSslRequest struct {
	InstanceId string                `json:"instance_id"`
	Body       *SwitchSslRequestBody `json:"body,omitempty"`
}

func (o SwitchSslRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SwitchSslRequest struct{}"
	}

	return strings.Join([]string{"SwitchSslRequest", string(data)}, " ")
}
