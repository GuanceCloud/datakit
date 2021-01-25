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
type CreateServerGroupRequestBody struct {
	ServerGroup *CreateServerGroupOption `json:"server_group"`
}

func (o CreateServerGroupRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateServerGroupRequestBody struct{}"
	}

	return strings.Join([]string{"CreateServerGroupRequestBody", string(data)}, " ")
}
