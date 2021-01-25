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
type CreateServerGroupRequest struct {
	Body *CreateServerGroupRequestBody `json:"body,omitempty"`
}

func (o CreateServerGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateServerGroupRequest struct{}"
	}

	return strings.Join([]string{"CreateServerGroupRequest", string(data)}, " ")
}
