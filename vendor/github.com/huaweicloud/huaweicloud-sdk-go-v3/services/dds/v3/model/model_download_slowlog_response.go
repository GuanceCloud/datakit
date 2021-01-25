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
type DownloadSlowlogResponse struct {
	// 具体信息。
	List *[]DownloadSlowlogResult `json:"list,omitempty"`
	// 查询状态。
	Status *string `json:"status,omitempty"`
	// 总记录数。
	TotalRecord    *string `json:"total_record,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DownloadSlowlogResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadSlowlogResponse struct{}"
	}

	return strings.Join([]string{"DownloadSlowlogResponse", string(data)}, " ")
}
