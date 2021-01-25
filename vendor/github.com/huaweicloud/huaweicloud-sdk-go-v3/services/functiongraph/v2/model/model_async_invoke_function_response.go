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
type AsyncInvokeFunctionResponse struct {
	// 请求ID。
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o AsyncInvokeFunctionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AsyncInvokeFunctionResponse struct{}"
	}

	return strings.Join([]string{"AsyncInvokeFunctionResponse", string(data)}, " ")
}
