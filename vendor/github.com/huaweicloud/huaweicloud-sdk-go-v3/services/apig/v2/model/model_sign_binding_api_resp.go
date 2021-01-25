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

type SignBindingApiResp struct {
	// API的发布编号
	PublishId *string `json:"publish_id,omitempty"`
	// API编号
	ApiId *string `json:"api_id,omitempty"`
	// 签名密钥的密钥
	SignSecret *string `json:"sign_secret,omitempty"`
	// API所属分组的名称
	GroupName *string `json:"group_name,omitempty"`
	// 签名密钥的编号
	SignId *string `json:"sign_id,omitempty"`
	// 签名密钥的key
	SignKey *string `json:"sign_key,omitempty"`
	// 绑定时间
	BindingTime *sdktime.SdkTime `json:"binding_time,omitempty"`
	// API所属环境的编号
	EnvId *string `json:"env_id,omitempty"`
	// API所属环境的名称
	EnvName *string `json:"env_name,omitempty"`
	// 签名密钥的名称
	SignName *string `json:"sign_name,omitempty"`
	// API类型
	ApiType *int32 `json:"api_type,omitempty"`
	// API名称
	ApiName *string `json:"api_name,omitempty"`
	// 绑定关系的ID
	Id *string `json:"id,omitempty"`
	// API描述
	ApiRemark *string `json:"api_remark,omitempty"`
}

func (o SignBindingApiResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SignBindingApiResp struct{}"
	}

	return strings.Join([]string{"SignBindingApiResp", string(data)}, " ")
}
