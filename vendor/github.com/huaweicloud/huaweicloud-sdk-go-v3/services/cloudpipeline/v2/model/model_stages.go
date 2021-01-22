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

// 流水线阶段执行信息
type Stages struct {
	// 阶段执行结果
	Result string `json:"result"`
	// 阶段执行状态
	Status string `json:"status"`
	// 阶段名字
	Name string `json:"name"`
	// 任务参数
	Parameters *interface{} `json:"parameters"`
	// 阶段顺序
	Order int32 `json:"order"`
	// 阶段类型
	DslMethod string `json:"dsl_method"`
	// 阶段显示名称
	DisplayName string `json:"display_name"`
}

func (o Stages) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Stages struct{}"
	}

	return strings.Join([]string{"Stages", string(data)}, " ")
}
