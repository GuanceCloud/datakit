/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type UserForCreation struct {
	// 数据库用户名称。 数据库帐号名称在1到32个字符之间，由小写字母、数字、中划线、或下划线组成，不能包含其他特殊字符。 - 若数据库版本为MySQL5.6和8.0，帐号长度为1～16个字符。 - 若数据库版本为MySQL5.7，帐号长度为1～32个字符。
	Name string `json:"name"`
	// 数据库帐号密码。  取值范围：  非空，由大小写字母、数字和特殊符号~!@#%^*-_=+?组成，长度8~32个字符，不能和数据库帐号“name”或“name”的逆序相同。  建议您输入高强度密码，以提高安全性，防止出现密码被暴力破解等安全风险。
	Password string `json:"password"`
}

func (o UserForCreation) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UserForCreation struct{}"
	}

	return strings.Join([]string{"UserForCreation", string(data)}, " ")
}
