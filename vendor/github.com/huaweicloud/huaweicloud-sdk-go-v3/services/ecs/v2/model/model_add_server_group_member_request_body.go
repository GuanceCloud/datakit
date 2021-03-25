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

// This is a auto create Body Object
type AddServerGroupMemberRequestBody struct {
	AddMember *ServerGroupMember `json:"add_member"`
}

func (o AddServerGroupMemberRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddServerGroupMemberRequestBody struct{}"
	}

	return strings.Join([]string{"AddServerGroupMemberRequestBody", string(data)}, " ")
}
