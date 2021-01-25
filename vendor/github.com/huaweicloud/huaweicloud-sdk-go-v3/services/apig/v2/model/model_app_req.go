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

type AppReq struct {
	// APP的名称。支持汉字，英文，数字，下划线，且只能以英文和汉字开头，3 ~ 64个字符。 > 中文字符必须为UTF-8或者unicode编码。
	Name string `json:"name"`
	// APP描述。字符长度不能大于255。 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// APP的key。支持英文，数字，“_”,“-”,且只能以英文或数字开头，8 ~ 64个字符。 > 只支持部分region。
	AppKey *string `json:"app_key,omitempty"`
	// 密钥。支持英文，数字，“_”,“-”,“_”,“!”,“@”,“#”,“$”,“%”且只能以英文或数字开头，8 ~ 64个字符。 > 只支持部分region。
	AppSecret *string `json:"app_secret,omitempty"`
}

func (o AppReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppReq struct{}"
	}

	return strings.Join([]string{"AppReq", string(data)}, " ")
}
