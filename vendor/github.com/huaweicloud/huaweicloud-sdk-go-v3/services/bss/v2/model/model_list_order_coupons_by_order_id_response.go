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

// Response Object
type ListOrderCouponsByOrderIdResponse struct {
	// |参数名称：符合条件的记录总数。| |参数的约束及描述：符合条件的记录总数。|
	Count *int32 `json:"count,omitempty"`
	// |参数名称：客户订单详情信息。具体请参见表 CustomerOrderV2| |参数约束以及描述：客户订单详情信息。具体请参见表 CustomerOrderV2|
	UserCoupons    *[]CouponInfoV2 `json:"user_coupons,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListOrderCouponsByOrderIdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOrderCouponsByOrderIdResponse struct{}"
	}

	return strings.Join([]string{"ListOrderCouponsByOrderIdResponse", string(data)}, " ")
}
