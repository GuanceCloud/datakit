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
type ShowOffSiteBackupPolicyRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
}

func (o ShowOffSiteBackupPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowOffSiteBackupPolicyRequest struct{}"
	}

	return strings.Join([]string{"ShowOffSiteBackupPolicyRequest", string(data)}, " ")
}
