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

// Request Object
type ListIssuedPartnerCouponsRequest struct {
	CouponId           *string `json:"coupon_id,omitempty"`
	CustomerId         *string `json:"customer_id,omitempty"`
	OrderId            *string `json:"order_id,omitempty"`
	CouponType         *int32  `json:"coupon_type,omitempty"`
	Status             *int32  `json:"status,omitempty"`
	CreateTimeBegin    *string `json:"create_time_begin,omitempty"`
	CreateTimeEnd      *string `json:"create_time_end,omitempty"`
	EffectiveTimeBegin *string `json:"effective_time_begin,omitempty"`
	EffectiveTimeEnd   *string `json:"effective_time_end,omitempty"`
	ExpireTimeBegin    *string `json:"expire_time_begin,omitempty"`
	ExpireTimeEnd      *string `json:"expire_time_end,omitempty"`
	Offset             *int32  `json:"offset,omitempty"`
	Limit              *int32  `json:"limit,omitempty"`
	IndirectPartnerId  *string `json:"indirect_partner_id,omitempty"`
}

func (o ListIssuedPartnerCouponsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIssuedPartnerCouponsRequest struct{}"
	}

	return strings.Join([]string{"ListIssuedPartnerCouponsRequest", string(data)}, " ")
}
