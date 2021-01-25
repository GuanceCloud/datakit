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

// Response Object
type ListSlowlogResponse struct {
	// 总数
	Count *int32 `json:"count,omitempty"`
	// 慢日志列表
	Slowlogs       *[]SlowlogItem `json:"slowlogs,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListSlowlogResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSlowlogResponse struct{}"
	}

	return strings.Join([]string{"ListSlowlogResponse", string(data)}, " ")
}
