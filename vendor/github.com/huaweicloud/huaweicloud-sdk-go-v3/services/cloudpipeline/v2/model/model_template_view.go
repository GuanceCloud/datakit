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

// 流水线创建状态响应体
type TemplateView struct {
	// 模板ID
	TemplateId string `json:"template_id"`
	// 模板名字
	TemplateName string `json:"template_name"`
	// 模板类型
	TemplateType string `json:"template_type"`
	// 模板编辑URL
	TemplateUrl string `json:"template_url"`
	// 用户ID
	UserId string `json:"user_id"`
	// 用户名字
	UserName string `json:"user_name"`
	// 租户ID
	DomainId string `json:"domain_id"`
	// 租户名字
	DomainName string `json:"domain_name"`
	// 是否内置模板
	IsBuildIn bool `json:"is_build_in"`
	// region
	Region string `json:"region"`
	// 项目ID
	ProjectId string `json:"project_id"`
	// 项目名字
	ProjectName string `json:"project_name"`
	// 是否关注
	IsWatch bool `json:"is_watch"`
	// 模板描述
	Description string `json:"description"`
	// 模板参数
	Parameter []TemplateParam `json:"parameter"`
	// 编排flow，map类型数据
	Flow *interface{} `json:"flow"`
	// 子任务states，map类型数据
	States *interface{} `json:"states"`
}

func (o TemplateView) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateView struct{}"
	}

	return strings.Join([]string{"TemplateView", string(data)}, " ")
}
