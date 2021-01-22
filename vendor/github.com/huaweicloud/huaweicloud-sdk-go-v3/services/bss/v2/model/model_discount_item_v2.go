/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type DiscountItemV2 struct {
	// |参数名称：折扣类型：200：促销产品折扣；300：促销折扣券；301：促销代金券；302：促销现金券；500：代理订购指定折扣；501：代理订购指定减免；502：代理订购指定一口价；600：折扣返利合同；601：渠道框架合同；602：专款专用合同；603：线下直签合同；604：电销授权合同；605：商务合同折扣；606：渠道商务合同折扣；607：合作伙伴授权折扣；609：订单调价折扣；700：促销折扣；800：充值帐户折扣；| |参数约束及描述：折扣类型：200：促销产品折扣；300：促销折扣券；301：促销代金券；302：促销现金券；500：代理订购指定折扣；501：代理订购指定减免；502：代理订购指定一口价；600：折扣返利合同；601：渠道框架合同；602：专款专用合同；603：线下直签合同；604：电销授权合同；605：商务合同折扣；606：渠道商务合同折扣；607：合作伙伴授权折扣；609：订单调价折扣；700：促销折扣；800：充值帐户折扣；|
	DiscountType *string `json:"discount_type,omitempty"`
	// |参数名称：折扣金额。| |参数的约束及描述：折扣金额。|
	DiscountAmount *float64 `json:"discount_amount,omitempty"`
}

func (o DiscountItemV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DiscountItemV2 struct{}"
	}

	return strings.Join([]string{"DiscountItemV2", string(data)}, " ")
}
