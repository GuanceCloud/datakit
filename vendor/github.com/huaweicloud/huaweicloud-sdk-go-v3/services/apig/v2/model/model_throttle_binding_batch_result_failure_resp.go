/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ThrottleBindingBatchResultFailureResp struct {
	// 解除绑定失败的API和流控策略绑定关系ID
	BindId *string `json:"bind_id,omitempty"`
	// 解除绑定失败的错误码
	ErrorCode *string `json:"error_code,omitempty"`
	// 解除绑定失败的错误信息
	ErrorMsg *string `json:"error_msg,omitempty"`
	// 解除绑定失败的API的ID
	ApiId *string `json:"api_id,omitempty"`
	// 解除绑定失败的API的名称
	ApiName *string `json:"api_name,omitempty"`
}

func (o ThrottleBindingBatchResultFailureResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleBindingBatchResultFailureResp struct{}"
	}

	return strings.Join([]string{"ThrottleBindingBatchResultFailureResp", string(data)}, " ")
}
