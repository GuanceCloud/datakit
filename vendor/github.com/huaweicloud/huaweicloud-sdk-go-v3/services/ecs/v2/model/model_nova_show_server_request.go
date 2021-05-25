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
type NovaShowServerRequest struct {
	ServerId            string  `json:"server_id"`
	OpenStackAPIVersion *string `json:"OpenStack-API-Version,omitempty"`
}

func (o NovaShowServerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaShowServerRequest struct{}"
	}

	return strings.Join([]string{"NovaShowServerRequest", string(data)}, " ")
}
