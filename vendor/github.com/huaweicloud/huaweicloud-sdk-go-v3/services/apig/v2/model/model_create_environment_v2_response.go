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

// Response Object
type CreateEnvironmentV2Response struct {
	// 创建时间
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 环境名称
	Name *string `json:"name,omitempty"`
	// 描述信息
	Remark *string `json:"remark,omitempty"`
	// 环境id
	Id             *string `json:"id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateEnvironmentV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEnvironmentV2Response struct{}"
	}

	return strings.Join([]string{"CreateEnvironmentV2Response", string(data)}, " ")
}
