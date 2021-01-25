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

type SetOffSiteBackupPolicyRequestBody struct {
	PolicyPara *OffSiteBackupPolicy `json:"policy_para"`
}

func (o SetOffSiteBackupPolicyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetOffSiteBackupPolicyRequestBody struct{}"
	}

	return strings.Join([]string{"SetOffSiteBackupPolicyRequestBody", string(data)}, " ")
}
