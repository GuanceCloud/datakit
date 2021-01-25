/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateDependencyResponse struct {
	// 依赖包ID。
	Id *string `json:"id,omitempty"`
	// 依赖包拥有者。
	Owner *string `json:"owner,omitempty"`
	// 依赖包在obs的存储地址。
	Link *string `json:"link,omitempty"`
	// 运行时语言。
	Runtime *string `json:"runtime,omitempty"`
	// 依赖包唯一标志。
	Etag *string `json:"etag,omitempty"`
	// 依赖包大小。
	Size *string `json:"size,omitempty"`
	// 依赖包名。
	Name *string `json:"name,omitempty"`
	// 依赖包描述。
	Description *string `json:"description,omitempty"`
	// 依赖包文件名。
	FileName       *string `json:"file_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateDependencyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateDependencyResponse struct{}"
	}

	return strings.Join([]string{"UpdateDependencyResponse", string(data)}, " ")
}
