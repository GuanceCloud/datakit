/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SetBackupPolicyRequestBody struct {
	BackupPolicy *BackupPolicy `json:"backup_policy"`
}

func (o SetBackupPolicyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetBackupPolicyRequestBody struct{}"
	}

	return strings.Join([]string{"SetBackupPolicyRequestBody", string(data)}, " ")
}
