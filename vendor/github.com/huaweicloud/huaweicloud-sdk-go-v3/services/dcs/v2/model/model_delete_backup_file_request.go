/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DeleteBackupFileRequest struct {
	BackupId   string `json:"backup_id"`
	InstanceId string `json:"instance_id"`
}

func (o DeleteBackupFileRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteBackupFileRequest struct{}"
	}

	return strings.Join([]string{"DeleteBackupFileRequest", string(data)}, " ")
}
