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

// 版本别名结构
type CreateVersionAliasRequestBody struct {
	// 要获取的别名名称。
	Name string `json:"name"`
	// 别名对应的版本名称。
	Version string `json:"version"`
	// 别名描述信息。
	Description *string `json:"description,omitempty"`
	// 灰度版本信息
	AdditionalVersionWeights map[string]int32 `json:"additional_version_weights,omitempty"`
}

func (o CreateVersionAliasRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVersionAliasRequestBody struct{}"
	}

	return strings.Join([]string{"CreateVersionAliasRequestBody", string(data)}, " ")
}
