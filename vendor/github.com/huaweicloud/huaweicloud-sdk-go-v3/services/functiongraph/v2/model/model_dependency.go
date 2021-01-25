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

// 函数依赖包结构。
type Dependency struct {
	// 依赖包属主的domainId。
	Owner string `json:"owner"`
	// 依赖包在OBS上的链接。
	Link string `json:"link"`
	// 依赖包语言类型，仅作为分类条件。
	Runtime string `json:"runtime"`
	// 依赖包的md5值
	Etag string `json:"etag"`
	// 依赖包大小。
	Size int64 `json:"size"`
	// 依赖包名称。
	Name string `json:"name"`
	// 依赖包描述。
	Description string `json:"description"`
	// 依赖包文件名，如果创建方式为zip时。
	FileName *string `json:"file_name,omitempty"`
}

func (o Dependency) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Dependency struct{}"
	}

	return strings.Join([]string{"Dependency", string(data)}, " ")
}
