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

// 子任务参数
type TemplateState struct {
	// 任务类型
	Type string `json:"type"`
	// 任务名字
	Name string `json:"name"`
	// 模板任务ID
	ModuleOrTemplateId string `json:"module_or_template_id"`
	// 模板任务名字
	ModuleOrTemplateName string `json:"module_or_template_name"`
	// 任务在流水线页面展示名字
	DisplayName string `json:"display_name"`
	// 流水线可挂载任务类型
	DslMethod string `json:"dsl_method"`
	// 任务参数,map类型数据
	Parameters *interface{} `json:"parameters"`
	// 是否手动执行
	IsManualExecution bool `json:"is_manual_execution"`
	// 任务参数是否校验
	JobParameterValidate bool `json:"job_parameter_validate"`
	// 是否显示代码仓URL
	IsShowCodehubUrl bool `json:"is_show_codehub_url"`
	// 是否执行
	IsExecute bool `json:"is_execute"`
	// 执行任务ID
	JobId string `json:"job_id"`
	// 执行任务名字
	JobName string `json:"job_name"`
	// 任务所属项目ID
	ProjectId string `json:"project_id"`
	// 执行方式
	ExecutionMode string `json:"execution_mode"`
}

func (o TemplateState) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateState struct{}"
	}

	return strings.Join([]string{"TemplateState", string(data)}, " ")
}
