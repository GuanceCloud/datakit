/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 授权结构体。
type AuthorizationVo struct {
	// 授权名称。
	Name string `json:"name"`
	// 仓库类型。 取值范围：github、gitlab、gitee、bitbucket、devcloud。
	RepoType string `json:"repo_type"`
	// 仓库地址。
	RepoHost string `json:"repo_host"`
	// 仓库主页。
	RepoHome string `json:"repo_home"`
	// 仓库用户名。
	RepoUser string `json:"repo_user"`
	// 头像。
	Avartar string `json:"avartar"`
	// 授权方式。
	TokenType string `json:"token_type"`
	// 创建时间。
	CreateTime int64 `json:"create_time"`
	// 修改时间。
	UpdateTime int64 `json:"update_time"`
	// 状态。
	Status int32 `json:"status"`
}

func (o AuthorizationVo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AuthorizationVo struct{}"
	}

	return strings.Join([]string{"AuthorizationVo", string(data)}, " ")
}
