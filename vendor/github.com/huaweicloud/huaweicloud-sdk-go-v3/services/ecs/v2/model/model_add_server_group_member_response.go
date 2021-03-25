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
type AddServerGroupMemberResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AddServerGroupMemberResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddServerGroupMemberResponse struct{}"
	}

	return strings.Join([]string{"AddServerGroupMemberResponse", string(data)}, " ")
}
