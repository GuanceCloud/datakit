/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DetachBaremetalServerVolumeRequest struct {
	ServerId     string `json:"server_id"`
	AttachmentId string `json:"attachment_id"`
}

func (o DetachBaremetalServerVolumeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DetachBaremetalServerVolumeRequest struct{}"
	}

	return strings.Join([]string{"DetachBaremetalServerVolumeRequest", string(data)}, " ")
}
