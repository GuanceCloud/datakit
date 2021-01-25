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

type AmountInfomationV2 struct {
	// |参数名称：费用项。具体请参见表 DiscountItemV2。| |参数约束以及描述：费用项。具体请参见表 DiscountItemV2。|
	Discounts *[]DiscountItemV2 `json:"discounts,omitempty"`
	// |参数名称：现金券金额，预留。| |参数的约束及描述：现金券金额，预留。|
	FlexipurchaseCouponAmount *float64 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：代金券金额。| |参数的约束及描述：代金券金额。|
	CouponAmount *float64 `json:"coupon_amount,omitempty"`
	// |参数名称：储值卡金额，预留。| |参数的约束及描述：储值卡金额，预留。|
	StoredCardAmount *float64 `json:"stored_card_amount,omitempty"`
	// |参数名称：手续费（仅退订订单存在）。| |参数的约束及描述：手续费（仅退订订单存在）。|
	CommissionAmount *float64 `json:"commission_amount,omitempty"`
	// |参数名称：消费金额（仅退订订单存在）。| |参数的约束及描述：消费金额（仅退订订单存在）。|
	ConsumedAmount *float64 `json:"consumed_amount,omitempty"`
}

func (o AmountInfomationV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AmountInfomationV2 struct{}"
	}

	return strings.Join([]string{"AmountInfomationV2", string(data)}, " ")
}
