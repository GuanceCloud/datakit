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

// 流水线参数详情
type Workflow struct {
	// 任务类型,list类型数据
	Parameter []PipelineParam `json:"parameter"`
	// 源码仓,list类型数据
	Source []Source `json:"source"`
	// 流水线名字
	Name string `json:"name"`
	// 项目ID
	ProjectId string `json:"project_id"`
	// 项目名字
	ProjectName string `json:"project_name"`
}

func (o Workflow) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Workflow struct{}"
	}

	return strings.Join([]string{"Workflow", string(data)}, " ")
}
