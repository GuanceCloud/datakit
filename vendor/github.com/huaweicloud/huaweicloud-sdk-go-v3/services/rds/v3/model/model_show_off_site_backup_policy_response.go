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

// Response Object
type ShowOffSiteBackupPolicyResponse struct {
	PolicyPara     *OffSiteBackupPolicy `json:"policy_para,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ShowOffSiteBackupPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowOffSiteBackupPolicyResponse struct{}"
	}

	return strings.Join([]string{"ShowOffSiteBackupPolicyResponse", string(data)}, " ")
}
