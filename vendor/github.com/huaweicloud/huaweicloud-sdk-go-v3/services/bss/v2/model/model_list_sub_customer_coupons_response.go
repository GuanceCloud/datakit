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
type ListSubCustomerCouponsResponse struct {
	// |参数名称：个数| |参数的约束及描述：个数|
	Count *int32 `json:"count,omitempty"`
	// |参数名称：优惠券记录。具体请参见表 IQueryUserCouponsResult。| |参数约束以及描述：优惠券记录。具体请参见表 IQueryUserCouponsResult。|
	UserCoupons    *[]IQueryUserCouponsResultV2 `json:"user_coupons,omitempty"`
	HttpStatusCode int                          `json:"-"`
}

func (o ListSubCustomerCouponsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubCustomerCouponsResponse struct{}"
	}

	return strings.Join([]string{"ListSubCustomerCouponsResponse", string(data)}, " ")
}
