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
type SetOffSiteBackupPolicyRequest struct {
	XLanguage  *string                            `json:"X-Language,omitempty"`
	InstanceId string                             `json:"instance_id"`
	Body       *SetOffSiteBackupPolicyRequestBody `json:"body,omitempty"`
}

func (o SetOffSiteBackupPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetOffSiteBackupPolicyRequest struct{}"
	}

	return strings.Join([]string{"SetOffSiteBackupPolicyRequest", string(data)}, " ")
}
