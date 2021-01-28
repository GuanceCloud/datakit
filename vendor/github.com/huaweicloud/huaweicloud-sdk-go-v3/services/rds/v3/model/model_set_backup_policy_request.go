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
type SetBackupPolicyRequest struct {
	XLanguage  *string                     `json:"X-Language,omitempty"`
	InstanceId string                      `json:"instance_id"`
	Body       *SetBackupPolicyRequestBody `json:"body,omitempty"`
}

func (o SetBackupPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetBackupPolicyRequest struct{}"
	}

	return strings.Join([]string{"SetBackupPolicyRequest", string(data)}, " ")
}
