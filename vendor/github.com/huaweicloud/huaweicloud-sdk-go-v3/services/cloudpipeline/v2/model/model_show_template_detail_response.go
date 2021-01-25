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

// Response Object
type ShowTemplateDetailResponse struct {
	// 模板ID
	TemplateId *string `json:"template_id,omitempty"`
	// 模板名字
	TemplateName *string `json:"template_name,omitempty"`
	// 模板类型
	TemplateType *string `json:"template_type,omitempty"`
	// 模板编辑URL
	TemplateUrl *string `json:"template_url,omitempty"`
	// 用户ID
	UserId *string `json:"user_id,omitempty"`
	// 用户名字
	UserName *string `json:"user_name,omitempty"`
	// 租户ID
	DomainId *string `json:"domain_id,omitempty"`
	// 租户名字
	DomainName *string `json:"domain_name,omitempty"`
	// 是否内置模板
	IsBuildIn *bool `json:"is_build_in,omitempty"`
	// region
	Region *string `json:"region,omitempty"`
	// 项目ID
	ProjectId *string `json:"project_id,omitempty"`
	// 项目名字
	ProjectName *string `json:"project_name,omitempty"`
	// 是否关注
	IsWatch *bool `json:"is_watch,omitempty"`
	// 模板描述
	Description *string `json:"description,omitempty"`
	// 模板参数
	Parameter *[]TemplateParam `json:"parameter,omitempty"`
	// 编排flow，map类型数据
	Flow *interface{} `json:"flow,omitempty"`
	// 子任务states，map类型数据
	States         *interface{} `json:"states,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ShowTemplateDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateDetailResponse struct{}"
	}

	return strings.Join([]string{"ShowTemplateDetailResponse", string(data)}, " ")
}
