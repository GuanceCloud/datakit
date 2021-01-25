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
type ListSubCustomerCouponsRequest struct {
	CouponId          *string `json:"coupon_id,omitempty"`
	OrderId           *string `json:"order_id,omitempty"`
	PromotionPlanId   *string `json:"promotion_plan_id,omitempty"`
	CouponType        *int32  `json:"coupon_type,omitempty"`
	Status            *int32  `json:"status,omitempty"`
	ActiveStartTime   *string `json:"active_start_time,omitempty"`
	ActiveEndTime     *string `json:"active_end_time,omitempty"`
	Offset            *int32  `json:"offset,omitempty"`
	Limit             *int32  `json:"limit,omitempty"`
	SourceId          *string `json:"source_id,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ListSubCustomerCouponsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubCustomerCouponsRequest struct{}"
	}

	return strings.Join([]string{"ListSubCustomerCouponsRequest", string(data)}, " ")
}
