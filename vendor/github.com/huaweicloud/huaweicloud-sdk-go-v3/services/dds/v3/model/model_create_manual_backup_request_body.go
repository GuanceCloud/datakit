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

type CreateManualBackupRequestBody struct {
	Backup *CreateManualBackupOption `json:"backup"`
}

func (o CreateManualBackupRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateManualBackupRequestBody struct{}"
	}

	return strings.Join([]string{"CreateManualBackupRequestBody", string(data)}, " ")
}
