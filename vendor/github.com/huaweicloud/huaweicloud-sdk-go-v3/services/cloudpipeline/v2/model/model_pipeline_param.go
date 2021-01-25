/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 流水线参数
type PipelineParam struct {
	// 流水线参数名字
	Name string `json:"name"`
	// 流水线参数值
	Value string `json:"value"`
	// 流水线参数描述
	Description string `json:"description"`
	// 流水线参数类型
	Paramtype string `json:"paramtype"`
	// 是否静态参数
	IsStatic bool `json:"is_static"`
	// 是否默认参数
	IsDefault bool `json:"is_default"`
}

func (o PipelineParam) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PipelineParam struct{}"
	}

	return strings.Join([]string{"PipelineParam", string(data)}, " ")
}
