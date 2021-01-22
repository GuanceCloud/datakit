/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

type ThrottleSpecialResp struct {
	// 特殊对象在流控时间内能够访问API的最大次数限制
	CallLimits *int32 `json:"call_limits,omitempty"`
	// 作用的APP名称
	AppName *string `json:"app_name,omitempty"`
	// 作用的APP或租户的名称
	ObjectName *string `json:"object_name,omitempty"`
	// 特殊对象的身份标识
	ObjectId *string `json:"object_id,omitempty"`
	// 流控策略编号
	ThrottleId *string `json:"throttle_id,omitempty"`
	// 设置时间
	ApplyTime *sdktime.SdkTime `json:"apply_time,omitempty"`
	// 特殊配置的编号
	Id *string `json:"id,omitempty"`
	// 作用的APP编号
	AppId *string `json:"app_id,omitempty"`
	// 特殊对象类型：APP、USER
	ObjectType *string `json:"object_type,omitempty"`
}

func (o ThrottleSpecialResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleSpecialResp struct{}"
	}

	return strings.Join([]string{"ThrottleSpecialResp", string(data)}, " ")
}
