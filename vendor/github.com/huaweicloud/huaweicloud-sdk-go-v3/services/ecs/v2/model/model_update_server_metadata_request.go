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
type UpdateServerMetadataRequest struct {
	ServerId string                           `json:"server_id"`
	Body     *UpdateServerMetadataRequestBody `json:"body,omitempty"`
}

func (o UpdateServerMetadataRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateServerMetadataRequest struct{}"
	}

	return strings.Join([]string{"UpdateServerMetadataRequest", string(data)}, " ")
}
