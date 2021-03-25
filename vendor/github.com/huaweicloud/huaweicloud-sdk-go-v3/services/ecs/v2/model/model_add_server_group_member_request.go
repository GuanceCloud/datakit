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
type AddServerGroupMemberRequest struct {
	ServerGroupId string                           `json:"server_group_id"`
	Body          *AddServerGroupMemberRequestBody `json:"body,omitempty"`
}

func (o AddServerGroupMemberRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddServerGroupMemberRequest struct{}"
	}

	return strings.Join([]string{"AddServerGroupMemberRequest", string(data)}, " ")
}
