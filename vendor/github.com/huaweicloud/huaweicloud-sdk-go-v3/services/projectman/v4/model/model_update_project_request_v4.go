/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 更新项目信息
type UpdateProjectRequestV4 struct {
	// 项目描述
	Description *string `json:"description,omitempty"`
	// 项目名
	ProjectName *string `json:"project_name,omitempty"`
}

func (o UpdateProjectRequestV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateProjectRequestV4 struct{}"
	}

	return strings.Join([]string{"UpdateProjectRequestV4", string(data)}, " ")
}
