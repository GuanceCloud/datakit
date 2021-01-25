/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type PropertiesInfo struct {
	// key
	Key *string `json:"key,omitempty"`
	// 默认值
	DefaultValue *string `json:"defaultValue,omitempty"`
	// 模板的描述信息
	Label *string `json:"label,omitempty"`
	// 类型 txet 或 select
	Type *string `json:"type,omitempty"`
	// 提示信息
	HelpText *string `json:"helpText,omitempty"`
	// 是否只读
	ReadOnly *bool `json:"readOnly,omitempty"`
	// 是否必填
	Required *bool `json:"required,omitempty"`
	// 正则校验类型
	RegType *string `json:"regType,omitempty"`
	// 正则表达式
	RegPattern *string `json:"regPattern,omitempty"`
	// 正则提示信息
	RegTip *string `json:"regTip,omitempty"`
	// 是否显示
	IsShow *bool `json:"isShow,omitempty"`
}

func (o PropertiesInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PropertiesInfo struct{}"
	}

	return strings.Join([]string{"PropertiesInfo", string(data)}, " ")
}
