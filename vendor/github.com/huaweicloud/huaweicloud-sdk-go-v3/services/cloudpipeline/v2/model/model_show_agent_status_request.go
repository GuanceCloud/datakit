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
type ShowAgentStatusRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
	AgentId   string  `json:"agent_id"`
}

func (o ShowAgentStatusRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowAgentStatusRequest struct{}"
	}

	return strings.Join([]string{"ShowAgentStatusRequest", string(data)}, " ")
}
