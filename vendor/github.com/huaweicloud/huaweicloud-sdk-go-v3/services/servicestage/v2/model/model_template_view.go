/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 模板参数。
type TemplateView struct {
	TemplateName *Template `json:"template_name,omitempty"`
	// 模板描述。
	TemplateDesc *string `json:"template_desc,omitempty"`
	// 模板类别。
	SourceType *string `json:"source_type,omitempty"`
	// 源码仓库URL
	SourceRepoUrl *string      `json:"source_repo_url,omitempty"`
	Runtime       *RuntimeType `json:"runtime,omitempty"`
}

func (o TemplateView) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateView struct{}"
	}

	return strings.Join([]string{"TemplateView", string(data)}, " ")
}
