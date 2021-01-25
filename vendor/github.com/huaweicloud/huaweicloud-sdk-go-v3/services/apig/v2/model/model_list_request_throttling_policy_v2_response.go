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

// Response Object
type ListRequestThrottlingPolicyV2Response struct {
	// 符合条件的流控策略的数量
	Total *int32 `json:"total,omitempty"`
	// 本次查询返回的列表长度
	Size *int32 `json:"size,omitempty"`
	// 本次查询到的流控策略列表
	Throttles      *[]ThrottleResp `json:"throttles,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListRequestThrottlingPolicyV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRequestThrottlingPolicyV2Response struct{}"
	}

	return strings.Join([]string{"ListRequestThrottlingPolicyV2Response", string(data)}, " ")
}
