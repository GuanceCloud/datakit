/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type AttachServerVolumeRequest struct {
	ServerId string                         `json:"server_id"`
	Body     *AttachServerVolumeRequestBody `json:"body,omitempty"`
}

func (o AttachServerVolumeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttachServerVolumeRequest struct{}"
	}

	return strings.Join([]string{"AttachServerVolumeRequest", string(data)}, " ")
}
