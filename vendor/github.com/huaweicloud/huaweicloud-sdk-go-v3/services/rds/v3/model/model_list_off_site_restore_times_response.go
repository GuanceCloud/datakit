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
type ListOffSiteRestoreTimesResponse struct {
	// 可恢复时间段列表。
	RestoreTime    *[]GetRestoreTimeResponseRestoreTime `json:"restore_time,omitempty"`
	HttpStatusCode int                                  `json:"-"`
}

func (o ListOffSiteRestoreTimesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOffSiteRestoreTimesResponse struct{}"
	}

	return strings.Join([]string{"ListOffSiteRestoreTimesResponse", string(data)}, " ")
}
