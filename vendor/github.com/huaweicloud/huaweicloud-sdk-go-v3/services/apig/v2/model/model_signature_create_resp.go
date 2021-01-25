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

type SignatureCreateResp struct {
	// 签名密钥的密钥
	SignSecret *string `json:"sign_secret,omitempty"`
	// 更新时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
	// 创建时间
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 签名密钥的名称
	Name *string `json:"name,omitempty"`
	// 签名密钥的编号
	Id *string `json:"id,omitempty"`
	// 签名密钥的key
	SignKey *string `json:"sign_key,omitempty"`
	// 签名密钥的类型
	SignType *string `json:"sign_type,omitempty"`
	// 绑定的API数量
	BindNum *int32 `json:"bind_num,omitempty"`
	// 绑定的自定义后端数量  暂不支持
	LdapiBindNum *int32 `json:"ldapi_bind_num,omitempty"`
}

func (o SignatureCreateResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SignatureCreateResp struct{}"
	}

	return strings.Join([]string{"SignatureCreateResp", string(data)}, " ")
}
