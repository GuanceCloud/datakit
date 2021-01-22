/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type AppSecretReq struct {
	// 密钥支持英文，数字，“_”,“-”,“_”,“!”,“@”,“#”,“$”,“%”,且只能以英文或数字开头，8 ~ 64个字符。用户自定义APP的密钥需要开启配额开关
	AppSecret *string `json:"app_secret,omitempty"`
}

func (o AppSecretReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppSecretReq struct{}"
	}

	return strings.Join([]string{"AppSecretReq", string(data)}, " ")
}
