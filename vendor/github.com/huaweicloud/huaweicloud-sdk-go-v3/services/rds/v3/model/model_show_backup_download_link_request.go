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
type ShowBackupDownloadLinkRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
	BackupId  string  `json:"backup_id"`
}

func (o ShowBackupDownloadLinkRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBackupDownloadLinkRequest struct{}"
	}

	return strings.Join([]string{"ShowBackupDownloadLinkRequest", string(data)}, " ")
}
