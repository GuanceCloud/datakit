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
type AsyncInvokeFunctionRequest struct {
	FunctionUrn string `json:"function_urn"`
	// 执行函数请求体，为json格式。
	Body map[string]interface{} `json:"body,omitempty"`
}

func (o AsyncInvokeFunctionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AsyncInvokeFunctionRequest struct{}"
	}

	return strings.Join([]string{"AsyncInvokeFunctionRequest", string(data)}, " ")
}
