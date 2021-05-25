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
type DeleteServerGroupMemberResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteServerGroupMemberResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteServerGroupMemberResponse struct{}"
	}

	return strings.Join([]string{"DeleteServerGroupMemberResponse", string(data)}, " ")
}
