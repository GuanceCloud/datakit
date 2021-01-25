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

type MemberListV4Members struct {
	// 租户id
	DomainId *string `json:"domain_id,omitempty"`
	// 租户名
	DomainName *string `json:"domain_name,omitempty"`
	// 用户id
	UserId *string `json:"user_id,omitempty"`
	// 用户名
	UserName *string `json:"user_name,omitempty"`
	// 创建人numId
	UserNumId *int32 `json:"user_num_id,omitempty"`
	// 成员角色 3 项目经理 4 开发人员 5 测试经理 6 测试人员 7 参与者 8 浏览者
	RoleId *int32 `json:"role_id,omitempty"`
	// 用户昵称
	NickName *string `json:"nick_name,omitempty"`
	// 用户角色
	RoleName *string `json:"role_name,omitempty"`
	// 用户类型 User iam用户 Federation 联邦账号
	UserType *string `json:"user_type,omitempty"`
	// 是否是禁用账号，1 禁用账号， 0非禁用账号
	Forbidden *int32 `json:"forbidden,omitempty"`
}

func (o MemberListV4Members) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MemberListV4Members struct{}"
	}

	return strings.Join([]string{"MemberListV4Members", string(data)}, " ")
}
