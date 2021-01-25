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
type ShowCurUserInfoResponse struct {
	// 租户id
	DomainId *string `json:"domain_id,omitempty"`
	// 租户名
	DomainName *string `json:"domain_name,omitempty"`
	// 用户数字id
	UserNumId *int32 `json:"user_num_id,omitempty"`
	// 用户id
	UserId *string `json:"user_id,omitempty"`
	// 用户名
	UserName *string `json:"user_name,omitempty"`
	// 用户昵称
	NickName *string `json:"nick_name,omitempty"`
	// 创建时间
	CreatedTime *string `json:"created_time,omitempty"`
	// 更新时间
	UpdatedTime *string `json:"updated_time,omitempty"`
	// 性别
	Gender *string `json:"gender,omitempty"`
	// 用户类型 User 云用户 Federation 联邦账号
	UserType       *string `json:"user_type,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowCurUserInfoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCurUserInfoResponse struct{}"
	}

	return strings.Join([]string{"ShowCurUserInfoResponse", string(data)}, " ")
}
