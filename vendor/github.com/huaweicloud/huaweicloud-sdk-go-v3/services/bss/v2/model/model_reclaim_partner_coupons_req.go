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

type ReclaimPartnerCouponsReq struct {
	// |参数名称：优惠券额度ID优惠券的类型跟随额度中的类型。| |参数约束及描述：优惠券额度ID优惠券的类型跟随额度中的类型。|
	CouponId string `json:"coupon_id"`
	// |参数名称：客户ID列表| |参数约束及描述：客户ID列表|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ReclaimPartnerCouponsReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimPartnerCouponsReq struct{}"
	}

	return strings.Join([]string{"ReclaimPartnerCouponsReq", string(data)}, " ")
}
