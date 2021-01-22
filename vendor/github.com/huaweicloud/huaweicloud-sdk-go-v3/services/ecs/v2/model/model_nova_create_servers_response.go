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
type NovaCreateServersResponse struct {
	Server         *NovaCreateServersResult `json:"server,omitempty"`
	HttpStatusCode int                      `json:"-"`
}

func (o NovaCreateServersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaCreateServersResponse struct{}"
	}

	return strings.Join([]string{"NovaCreateServersResponse", string(data)}, " ")
}
