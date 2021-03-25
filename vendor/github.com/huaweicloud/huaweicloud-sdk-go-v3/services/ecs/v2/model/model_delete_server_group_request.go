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
type DeleteServerGroupRequest struct {
	ServerGroupId string `json:"server_group_id"`
}

func (o DeleteServerGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteServerGroupRequest struct{}"
	}

	return strings.Join([]string{"DeleteServerGroupRequest", string(data)}, " ")
}
