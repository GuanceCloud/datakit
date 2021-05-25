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
type DeleteServerMetadataRequest struct {
	Key      string `json:"key"`
	ServerId string `json:"server_id"`
}

func (o DeleteServerMetadataRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteServerMetadataRequest struct{}"
	}

	return strings.Join([]string{"DeleteServerMetadataRequest", string(data)}, " ")
}
