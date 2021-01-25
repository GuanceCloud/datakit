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

type CreateDatabaseRoleRequestBody struct {
	// 创建角色名称。 - 长度为1~64位，可以包含大写字母（A~Z）、小写字母（a~z）、数字（0~9）、中划线、下划线和点。
	RoleName string `json:"role_name"`
	// 角色所在的数据库名称，默认admin。 - 长度为1~64位，可以包含大写字母（A~Z）、小写字母（a~z）、数字（0~9）、下划线。
	DbName *string      `json:"db_name,omitempty"`
	Roles  *RolesOption `json:"roles,omitempty"`
}

func (o CreateDatabaseRoleRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatabaseRoleRequestBody struct{}"
	}

	return strings.Join([]string{"CreateDatabaseRoleRequestBody", string(data)}, " ")
}
