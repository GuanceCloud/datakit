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

// Response Object
type ShowAgentStatusResponse struct {
	// Agent状态
	Status *string `json:"status,omitempty"`
	// AgentID
	AgentId        *string `json:"agent_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowAgentStatusResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowAgentStatusResponse struct{}"
	}

	return strings.Join([]string{"ShowAgentStatusResponse", string(data)}, " ")
}
