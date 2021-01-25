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

type ApiPolicyMockBase struct {
	// 返回结果
	ResultContent *string `json:"result_content,omitempty"`
}

func (o ApiPolicyMockBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyMockBase struct{}"
	}

	return strings.Join([]string{"ApiPolicyMockBase", string(data)}, " ")
}
