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

type ComponentView struct {
	// 组件ID。
	Id *string `json:"id,omitempty"`
	// 应用ID。
	ApplicationId *string `json:"application_id,omitempty"`
	// 应用组件名称。
	Name        *string               `json:"name,omitempty"`
	Runtime     *RuntimeType          `json:"runtime,omitempty"`
	Category    *ComponentCategory    `json:"category,omitempty"`
	SubCategory *ComponentSubCategory `json:"sub_category,omitempty"`
	// 组件描述。
	Description *string `json:"description,omitempty"`
	// 取值0或1。  0：表示正常状态。  1：表示正在删除。
	Status *int32        `json:"status,omitempty"`
	Source *SourceObject `json:"source,omitempty"`
	// 创建人。
	Creator *string `json:"creator,omitempty"`
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 修改时间。
	UpdateTime *int64 `json:"update_time,omitempty"`
}

func (o ComponentView) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ComponentView struct{}"
	}

	return strings.Join([]string{"ComponentView", string(data)}, " ")
}
