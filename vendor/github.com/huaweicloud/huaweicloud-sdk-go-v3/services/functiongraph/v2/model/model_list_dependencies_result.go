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

// 依赖包信息。
type ListDependenciesResult struct {
	// 依赖包ID。
	Id string `json:"id"`
	// 依赖包拥有者。
	Owner string `json:"owner"`
	// 依赖包在obs的存储地址。
	Link string `json:"link"`
	// 运行时语言。
	Runtime string `json:"runtime"`
	// 依赖包唯一标志。
	Etag string `json:"etag"`
	// 依赖包大小。
	Size string `json:"size"`
	// 依赖包名。
	Name string `json:"name"`
	// 依赖包描述。
	Description *string `json:"description,omitempty"`
	// 依赖包文件名。
	FileName *string `json:"file_name,omitempty"`
}

func (o ListDependenciesResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDependenciesResult struct{}"
	}

	return strings.Join([]string{"ListDependenciesResult", string(data)}, " ")
}
