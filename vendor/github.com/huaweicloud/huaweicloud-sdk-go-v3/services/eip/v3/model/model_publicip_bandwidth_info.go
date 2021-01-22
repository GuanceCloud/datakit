/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 公网IP绑定的带宽信息
type PublicipBandwidthInfo struct {
	// 带宽ID
	Id *string `json:"id,omitempty"`
	// 功能描述：带宽大小 取值范围：默认5Mbit/s~2000Mbit/s
	Size *int32 `json:"size,omitempty"`
	// 功能说明：带宽类型,标识是否是共享带宽 取值范围：PER，WHOLE。   PER：独享带宽   WHOLE：共享带宽 约束：其中IPv6暂不支持WHOLE类型带宽。
	ShareType *string `json:"share_type,omitempty"`
	// 功能说明：按流量计费还是按带宽计费 取值范围： bandwidth：按带宽计费 traffic：按流量计费 95peak_plus：按增强型95计费
	ChargeMode *string `json:"charge_mode,omitempty"`
	// 功能说明：带宽名称 取值范围：1-64个字符,支持数字、字母、中文、_(下划线)、-(中划线)、.(点)
	Name *string `json:"name,omitempty"`
	// 功能说明：账单信息。如果billinginfo不为空，说明是包周期的带宽
	BillingInfo *string `json:"billing_info,omitempty"`
}

func (o PublicipBandwidthInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublicipBandwidthInfo struct{}"
	}

	return strings.Join([]string{"PublicipBandwidthInfo", string(data)}, " ")
}
