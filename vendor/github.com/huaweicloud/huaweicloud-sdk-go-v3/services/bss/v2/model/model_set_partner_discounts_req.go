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

type SetPartnerDiscountsReq struct {
	// |参数名称：二级经销商ID| |参数约束及描述：一级经销商给二级经销商的子客户设置折扣时需要携带这个字段。|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
	// |参数名称：客户折扣信息列表，最大支持10个。| |参数约束以及描述：客户折扣信息列表，最大支持10个。|
	SubCustomerDiscounts []SetSubCustomerDiscountV2 `json:"sub_customer_discounts"`
}

func (o SetPartnerDiscountsReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetPartnerDiscountsReq struct{}"
	}

	return strings.Join([]string{"SetPartnerDiscountsReq", string(data)}, " ")
}
