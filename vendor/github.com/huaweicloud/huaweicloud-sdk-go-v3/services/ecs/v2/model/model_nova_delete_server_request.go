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
type NovaDeleteServerRequest struct {
	ServerId string `json:"server_id"`
}

func (o NovaDeleteServerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaDeleteServerRequest struct{}"
	}

	return strings.Join([]string{"NovaDeleteServerRequest", string(data)}, " ")
}
