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

// Request Object
type DeleteManualBackupRequest struct {
	BackupId string `json:"backup_id"`
}

func (o DeleteManualBackupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteManualBackupRequest struct{}"
	}

	return strings.Join([]string{"DeleteManualBackupRequest", string(data)}, " ")
}
