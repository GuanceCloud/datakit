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
type ListDatabaseUsersResponse struct {
	// 数据库用户信息。
	Users *string `json:"users,omitempty"`
	// 数据库用户总数。
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListDatabaseUsersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDatabaseUsersResponse struct{}"
	}

	return strings.Join([]string{"ListDatabaseUsersResponse", string(data)}, " ")
}
