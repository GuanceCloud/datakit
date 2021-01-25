/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type GetCursorResponse struct {
	// 数据游标。  取值范围：1~512个字符。  说明：  数据游标有效期为5分钟。
	PartitionCursor *string `json:"partition_cursor,omitempty"`
	HttpStatusCode  int     `json:"-"`
}

func (o GetCursorResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetCursorResponse struct{}"
	}

	return strings.Join([]string{"GetCursorResponse", string(data)}, " ")
}
