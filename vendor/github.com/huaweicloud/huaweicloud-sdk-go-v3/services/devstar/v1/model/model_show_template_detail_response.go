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

// Response Object
type ShowTemplateDetailResponse struct {
	// 模板的id
	Id *string `json:"id,omitempty"`
	// 模板的名称
	Title *string `json:"title,omitempty"`
	// 模板的描述信息
	Description *string `json:"description,omitempty"`
	// 模板关联的region host id
	RegionId *string `json:"region_id,omitempty"`
	// 模板关联的repo id
	RepostoryId *string `json:"repostory_id,omitempty"`
	// 模板https下载路径
	CodeUrl *string `json:"code_url,omitempty"`
	// 模板ssh下载路径
	SshUrl *string `json:"ssh_url,omitempty"`
	// 项目id
	ProjectUuid *string `json:"project_uuid,omitempty"`
	// 模板状态
	Status         *int32            `json:"status,omitempty"`
	Properties     *[]PropertiesInfo `json:"properties,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ShowTemplateDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateDetailResponse struct{}"
	}

	return strings.Join([]string{"ShowTemplateDetailResponse", string(data)}, " ")
}
