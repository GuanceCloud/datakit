/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CheckUserIdentityResponse struct {
	// |参数名称：返回码| |参数的约束及描述：该参数必填，且只允许字符串|
	ErrorCode *string `json:"error_code,omitempty"`
	// |参数名称：返回码描述| |参数的约束及描述：该参数必填，且只允许字符串|
	ErrorMsg *string `json:"error_msg,omitempty"`
	// |参数名称：是否可以继续注册| |参数的约束及描述：该参数非必填，且只允许字符串,available: 该登录名称/手机号/邮箱可以继续注册,used_by_user: 该登录名称/手机号/邮箱不可以继续注册|
	CheckResult    *string `json:"check_result,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CheckUserIdentityResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CheckUserIdentityResponse struct{}"
	}

	return strings.Join([]string{"CheckUserIdentityResponse", string(data)}, " ")
}
