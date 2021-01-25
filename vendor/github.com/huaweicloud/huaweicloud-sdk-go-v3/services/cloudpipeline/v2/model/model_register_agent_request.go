/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type RegisterAgentRequest struct {
	Body *SlaveRegister `json:"body,omitempty"`
}

func (o RegisterAgentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RegisterAgentRequest struct{}"
	}

	return strings.Join([]string{"RegisterAgentRequest", string(data)}, " ")
}
