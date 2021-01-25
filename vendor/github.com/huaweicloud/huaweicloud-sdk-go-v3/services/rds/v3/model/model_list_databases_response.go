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
type ListDatabasesResponse struct {
	// 数据库信息。
	Databases *[]DatabaseForList `json:"databases,omitempty"`
	// 总数。
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListDatabasesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDatabasesResponse struct{}"
	}

	return strings.Join([]string{"ListDatabasesResponse", string(data)}, " ")
}
