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

type ThrottleSpecialUpdateReq struct {
	// 流控时间内特殊对象能够访问API的最大次数限制
	CallLimits int32 `json:"call_limits"`
}

func (o ThrottleSpecialUpdateReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleSpecialUpdateReq struct{}"
	}

	return strings.Join([]string{"ThrottleSpecialUpdateReq", string(data)}, " ")
}
