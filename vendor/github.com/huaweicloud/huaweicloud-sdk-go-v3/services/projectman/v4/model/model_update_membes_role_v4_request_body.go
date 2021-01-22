/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type UpdateMembesRoleV4RequestBody struct {
	// 成员角色 3 项目经理 4 开发人员 5 测试经理 6 测试人员 7 参与者 8 浏览者
	RoleId int32 `json:"role_id"`
	// 用户id
	UserIds []string `json:"user_ids"`
}

func (o UpdateMembesRoleV4RequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateMembesRoleV4RequestBody struct{}"
	}

	return strings.Join([]string{"UpdateMembesRoleV4RequestBody", string(data)}, " ")
}
