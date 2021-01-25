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
type ResizeInstanceVolumeRequest struct {
	InstanceId string                           `json:"instance_id"`
	Body       *ResizeInstanceVolumeRequestBody `json:"body,omitempty"`
}

func (o ResizeInstanceVolumeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceVolumeRequest struct{}"
	}

	return strings.Join([]string{"ResizeInstanceVolumeRequest", string(data)}, " ")
}
