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

// Request Object
type InvokeFunctionRequest struct {
	FunctionUrn        string  `json:"function_urn"`
	XCffLogType        *string `json:"X-Cff-Log-Type,omitempty"`
	XCFFRequestVersion *string `json:"X-CFF-Request-Version,omitempty"`
	// 执行函数请求体，为json格式。
	Body map[string]interface{} `json:"body,omitempty"`
}

func (o InvokeFunctionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InvokeFunctionRequest struct{}"
	}

	return strings.Join([]string{"InvokeFunctionRequest", string(data)}, " ")
}
