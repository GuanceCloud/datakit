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

type ConfigurationForUpdate struct {
	// 参数模板名称。最长64个字符，只允许大写字母、小写字母、数字、和“-_.”特殊字符。
	Name *string `json:"name,omitempty"`
	// 参数模板描述。最长256个字符，不支持>!<\"&'=特殊字符。默认为空。
	Description *string `json:"description,omitempty"`
	// 参数值对象，用户基于默认参数模板自定义的参数值。为空时不修改参数值。
	Values map[string]string `json:"values,omitempty"`
}

func (o ConfigurationForUpdate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConfigurationForUpdate struct{}"
	}

	return strings.Join([]string{"ConfigurationForUpdate", string(data)}, " ")
}
