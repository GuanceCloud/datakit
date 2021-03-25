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
type DeleteServerGroupMemberRequest struct {
	ServerGroupId string                              `json:"server_group_id"`
	Body          *DeleteServerGroupMemberRequestBody `json:"body,omitempty"`
}

func (o DeleteServerGroupMemberRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteServerGroupMemberRequest struct{}"
	}

	return strings.Join([]string{"DeleteServerGroupMemberRequest", string(data)}, " ")
}
