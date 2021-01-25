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

type AppInfoWithBindNumResp struct {
	// APP的创建者 - USER：用户自行创建 - MARKET：云市场分配
	Creator *string `json:"creator,omitempty"`
	// 更新时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
	// APP的key
	AppKey *string `json:"app_key,omitempty"`
	// 名称
	Name *string `json:"name,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 编号
	Id *string `json:"id,omitempty"`
	// 密钥
	AppSecret *string `json:"app_secret,omitempty"`
	// 注册时间
	RegisterTime *sdktime.SdkTime `json:"register_time,omitempty"`
	// 状态
	Status *int32 `json:"status,omitempty"`
	// 绑定的API数量
	BindNum *int32 `json:"bind_num,omitempty"`
}

func (o AppInfoWithBindNumResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppInfoWithBindNumResp struct{}"
	}

	return strings.Join([]string{"AppInfoWithBindNumResp", string(data)}, " ")
}
