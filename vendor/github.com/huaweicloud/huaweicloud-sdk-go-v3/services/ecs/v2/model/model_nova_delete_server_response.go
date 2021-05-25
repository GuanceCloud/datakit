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

// Response Object
type NovaDeleteServerResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NovaDeleteServerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaDeleteServerResponse struct{}"
	}

	return strings.Join([]string{"NovaDeleteServerResponse", string(data)}, " ")
}
