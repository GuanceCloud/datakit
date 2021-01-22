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

type CreateFunctionVersionRequestBody struct {
	// md5键值
	Digest *string `json:"digest,omitempty"`
	// 发布版本名称
	Version *string `json:"version,omitempty"`
	// 发布版本描述
	Description *string `json:"description,omitempty"`
}

func (o CreateFunctionVersionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateFunctionVersionRequestBody struct{}"
	}

	return strings.Join([]string{"CreateFunctionVersionRequestBody", string(data)}, " ")
}
