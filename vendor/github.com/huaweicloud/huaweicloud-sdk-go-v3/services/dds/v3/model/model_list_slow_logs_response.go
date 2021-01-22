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

// Response Object
type ListSlowLogsResponse struct {
	// 具体信息。
	SlowLogList *[]SlowlogResult `json:"slow_log_list,omitempty"`
	// 数据库版本总记录数。
	TotalRecord    *string `json:"total_record,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListSlowLogsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSlowLogsResponse struct{}"
	}

	return strings.Join([]string{"ListSlowLogsResponse", string(data)}, " ")
}
