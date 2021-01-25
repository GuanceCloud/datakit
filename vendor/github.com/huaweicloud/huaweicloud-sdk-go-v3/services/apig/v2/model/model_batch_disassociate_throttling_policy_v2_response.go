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
type BatchDisassociateThrottlingPolicyV2Response struct {
	// 成功解除绑定的API和流控策略绑定关系的数量
	SuccessCount *int32 `json:"success_count,omitempty"`
	// 解除绑定失败的API和流控绑定关系及错误信息
	Failure        *[]ThrottleBindingBatchResultFailureResp `json:"failure,omitempty"`
	HttpStatusCode int                                      `json:"-"`
}

func (o BatchDisassociateThrottlingPolicyV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDisassociateThrottlingPolicyV2Response struct{}"
	}

	return strings.Join([]string{"BatchDisassociateThrottlingPolicyV2Response", string(data)}, " ")
}
