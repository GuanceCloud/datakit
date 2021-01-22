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

// 扩展参数，用于包周期资源申请
type UpdatePrePaidBandwidthExtendParamOption struct {
	// 功能说明：下单订购后，是否自动从客户的账户中支付，而不需要客户手动去进行支付；系统默认是“非自动支付”。  取值范围：  true：是（自动支付）  false：否（默认值，需要客户手动去支付）  约束：自动支付时，只能使用账户的现金支付；如果要使用代金券，请选择不自动支付，然后在用户费用中心，选择代金券支付。
	IsAutoPay *bool `json:"is_auto_pay,omitempty"`
}

func (o UpdatePrePaidBandwidthExtendParamOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePrePaidBandwidthExtendParamOption struct{}"
	}

	return strings.Join([]string{"UpdatePrePaidBandwidthExtendParamOption", string(data)}, " ")
}
