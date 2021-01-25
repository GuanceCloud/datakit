/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListRetentionHistoriesResponse struct {
	// 镜像老化日志
	RetentionLog *[]RetentionLog `json:"retention_log,omitempty"`
	// 总个数
	Total          *int32  `json:"total,omitempty"`
	ContentRange   *string `json:"Content-Range,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListRetentionHistoriesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRetentionHistoriesResponse struct{}"
	}

	return strings.Join([]string{"ListRetentionHistoriesResponse", string(data)}, " ")
}
