/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type SetAuditlogPolicyRequest struct {
	XLanguage  *string                       `json:"X-Language,omitempty"`
	InstanceId string                        `json:"instance_id"`
	Body       *SetAuditlogPolicyRequestBody `json:"body,omitempty"`
}

func (o SetAuditlogPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetAuditlogPolicyRequest struct{}"
	}

	return strings.Join([]string{"SetAuditlogPolicyRequest", string(data)}, " ")
}
