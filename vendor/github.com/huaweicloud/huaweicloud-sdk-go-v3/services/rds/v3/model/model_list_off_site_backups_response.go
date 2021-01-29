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
type ListOffSiteBackupsResponse struct {
	// 跨区域备份信息。
	OffsiteBackups *[]OffSiteBackupForList `json:"offsite_backups,omitempty"`
	// 总记录数。
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListOffSiteBackupsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOffSiteBackupsResponse struct{}"
	}

	return strings.Join([]string{"ListOffSiteBackupsResponse", string(data)}, " ")
}
