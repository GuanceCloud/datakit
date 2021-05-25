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
type DeleteServerGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteServerGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteServerGroupResponse struct{}"
	}

	return strings.Join([]string{"DeleteServerGroupResponse", string(data)}, " ")
}
