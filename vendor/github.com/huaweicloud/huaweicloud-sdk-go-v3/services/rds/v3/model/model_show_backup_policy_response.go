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
type ShowBackupPolicyResponse struct {
	BackupPolicy   *BackupPolicy `json:"backup_policy,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ShowBackupPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBackupPolicyResponse struct{}"
	}

	return strings.Join([]string{"ShowBackupPolicyResponse", string(data)}, " ")
}
