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
type ShowBackupDownloadLinkRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	BackupId   string  `json:"backup_id"`
}

func (o ShowBackupDownloadLinkRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBackupDownloadLinkRequest struct{}"
	}

	return strings.Join([]string{"ShowBackupDownloadLinkRequest", string(data)}, " ")
}
