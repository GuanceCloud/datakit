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

type ThrottleBindingBatchDelete struct {
	// 需要解除绑定的API和流控策略绑定关系ID列表
	ThrottleBindings *[]string `json:"throttle_bindings,omitempty"`
}

func (o ThrottleBindingBatchDelete) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleBindingBatchDelete struct{}"
	}

	return strings.Join([]string{"ThrottleBindingBatchDelete", string(data)}, " ")
}
