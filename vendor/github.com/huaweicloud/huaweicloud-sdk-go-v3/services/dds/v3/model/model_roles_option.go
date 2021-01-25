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

type RolesOption struct {
	// 被继承角色所在数据库名称。 - 长度为1~64位，可以包含大写字母（A~Z）、小写字母（a~z）、数字（0~9）、下划线。
	RoleDbName string `json:"role_db_name"`
	// 被继承角色的名称。 - 长度为1~64位，可以包含大写字母（A~Z）、小写字母（a~z）、数字（0~9）、中划线、下划线和点。
	RoleName string `json:"role_name"`
}

func (o RolesOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RolesOption struct{}"
	}

	return strings.Join([]string{"RolesOption", string(data)}, " ")
}
