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
type UpdateInstancePortRequest struct {
	InstanceId string                 `json:"instance_id"`
	Body       *UpdatePortRequestBody `json:"body,omitempty"`
}

func (o UpdateInstancePortRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstancePortRequest struct{}"
	}

	return strings.Join([]string{"UpdateInstancePortRequest", string(data)}, " ")
}
