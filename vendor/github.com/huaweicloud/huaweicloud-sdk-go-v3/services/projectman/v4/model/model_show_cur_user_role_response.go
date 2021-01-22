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

// Response Object
type ShowCurUserRoleResponse struct {
	// 成员角色 -1 项目创建者 3 项目经理 4 开发人员 5 测试经理 6 测试人员 7 参与者 8 浏览
	UserRole       *int32 `json:"user_role,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ShowCurUserRoleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCurUserRoleResponse struct{}"
	}

	return strings.Join([]string{"ShowCurUserRoleResponse", string(data)}, " ")
}
