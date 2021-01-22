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

type EnvVariableReq struct {
	// 变量值支持英文字母、数字、英文格式的下划线、中划线，斜线（/）、点、冒号，1 ~ 255个字符。
	VariableValue string `json:"variable_value"`
	// 环境编号
	EnvId string `json:"env_id"`
	// API分组编号
	GroupId string `json:"group_id"`
	// 变量名，支持英文字母、数字、英文格式的下划线、中划线，必须以英文字母开头, 3 ~ 32个字符。在API定义中等于#Name的值#部分（区分大小写），发布到环境里的API被变量值换。 > 中文字符必须为UTF-8或者unicode编码。
	VariableName string `json:"variable_name"`
}

func (o EnvVariableReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnvVariableReq struct{}"
	}

	return strings.Join([]string{"EnvVariableReq", string(data)}, " ")
}
