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
type ShowDetailsOfApiGroupV2Response struct {
	// 分组上绑定的独立域名列表
	UrlDomains *[]UrlDomainsResp `json:"url_domains,omitempty"`
	// 流控时长内分组下的API的总访问次数限制，默认不限，请根据服务的负载能力自行设置  暂不支持
	CallLimits *int32 `json:"call_limits,omitempty"`
	// 最近修改时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
	// API分组名称
	Name *string `json:"name,omitempty"`
	// 流控的时间单位  暂不支持
	TimeUnit *string `json:"time_unit,omitempty"`
	// 是否已上架云市场： - 1：已上架 - 2：未上架 - 3：审核中
	OnSellStatus *int32 `json:"on_sell_status,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 系统默认分配的子域名
	SlDomain *string `json:"sl_domain,omitempty"`
	// 系统默认分配的子域名列表
	SlDomains *[]string `json:"sl_domains,omitempty"`
	// 编号
	Id *string `json:"id,omitempty"`
	// 流控时长  暂不支持
	TimeInterval *int32 `json:"time_interval,omitempty"`
	// 创建时间
	RegisterTime *sdktime.SdkTime `json:"register_time,omitempty"`
	// 状态
	Status *int32 `json:"status,omitempty"`
	// 是否为默认分组
	IsDefault      *int32 `json:"is_default,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ShowDetailsOfApiGroupV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowDetailsOfApiGroupV2Response struct{}"
	}

	return strings.Join([]string{"ShowDetailsOfApiGroupV2Response", string(data)}, " ")
}
