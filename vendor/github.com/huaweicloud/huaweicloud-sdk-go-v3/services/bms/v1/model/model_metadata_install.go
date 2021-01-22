/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// metadata字段数据结构说明
type MetadataInstall struct {
	// 重装裸金属服务器过程中注入Linux镜像root密码，用户自定义初始化密码。注：修改密码脚本需经Base64编码。建议密码复杂度如下：长度为8-26位。密码至少必须包含大写字母（A-Z）、小写字母（a-z）、数字（0-9）和特殊字符（!@$%^-_=+[{}]:,./?）中的三种
	UserData *string `json:"user_data,omitempty"`
}

func (o MetadataInstall) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MetadataInstall struct{}"
	}

	return strings.Join([]string{"MetadataInstall", string(data)}, " ")
}
