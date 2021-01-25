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

// Response Object
type CreateProjectV4Response struct {
	// 项目数字id
	ProjectNumId *int32 `json:"project_num_id,omitempty"`
	// 项目id
	ProjectId *string `json:"project_id,omitempty"`
	// 项目名
	ProjectName *string `json:"project_name,omitempty"`
	// 项目描述
	Description *string `json:"description,omitempty"`
	// 项目类型
	ProjectType *string `json:"project_type,omitempty"`
	// 创建者的数字id
	UserNumId      *int32 `json:"user_num_id,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o CreateProjectV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateProjectV4Response struct{}"
	}

	return strings.Join([]string{"CreateProjectV4Response", string(data)}, " ")
}
