/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 挂载用户信息。
type MountUser struct {
	// 用户ID(-1~65534的非0整数)
	UserId string `json:"user_id"`
	// 用户组ID(-1~65534的非0整数)
	UserGroupId string `json:"user_group_id"`
}

func (o MountUser) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MountUser struct{}"
	}

	return strings.Join([]string{"MountUser", string(data)}, " ")
}
