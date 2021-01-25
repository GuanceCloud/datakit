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

type AmountInfomation struct {
	// |参数名称：费用项。| |参数约束及描述： 费用项。|
	Discounts *[]DiscountEntry `json:"discounts,omitempty"`
	// |参数名称：现金券金额。| |参数的约束及描述：现金券金额。|
	FlexipurchaseCouponAmount *float64 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：代金券金额。| |参数的约束及描述：代金券金额。|
	CouponAmount *float64 `json:"coupon_amount,omitempty"`
	// |参数名称：储值卡金额。| |参数的约束及描述：储值卡金额。|
	StoredCardAmount *float64 `json:"stored_card_amount,omitempty"`
}

func (o AmountInfomation) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AmountInfomation struct{}"
	}

	return strings.Join([]string{"AmountInfomation", string(data)}, " ")
}
