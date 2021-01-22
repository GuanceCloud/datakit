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

// Response Object
type CreateComponentResponse struct {
	// 应用组件ID。
	Id *string `json:"id,omitempty"`
	// 应用组件名称
	Name *string `json:"name,omitempty"`
	// 取值0或1。  0：表示正常状态。  1：表示正在删除。
	Status      *int32                `json:"status,omitempty"`
	Runtime     *RuntimeType          `json:"runtime,omitempty"`
	Category    *ComponentCategory    `json:"category,omitempty"`
	SubCategory *ComponentSubCategory `json:"sub_category,omitempty"`
	// 描述。
	Description *string `json:"description,omitempty"`
	// 项目ID。
	ProjectId *string `json:"project_id,omitempty"`
	// 应用ID。
	ApplicationId *string       `json:"application_id,omitempty"`
	Source        *SourceObject `json:"source,omitempty"`
	Build         *BuildInfo    `json:"build,omitempty"`
	// 流水线Id列表，最多10个。
	PipelineIds *[]string `json:"pipeline_ids,omitempty"`
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 修改时间。
	UpdateTime     *int64 `json:"update_time,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o CreateComponentResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateComponentResponse struct{}"
	}

	return strings.Join([]string{"CreateComponentResponse", string(data)}, " ")
}
