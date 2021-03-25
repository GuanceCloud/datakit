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
type UpdateServerResponse struct {
	Server         *UpdateServerResult `json:"server,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o UpdateServerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateServerResponse struct{}"
	}

	return strings.Join([]string{"UpdateServerResponse", string(data)}, " ")
}
