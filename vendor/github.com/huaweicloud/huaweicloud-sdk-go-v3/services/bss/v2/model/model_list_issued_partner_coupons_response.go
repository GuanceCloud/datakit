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
type ListIssuedPartnerCouponsResponse struct {
	// |参数名称：个数| |参数的约束及描述：个数|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：优惠券记录。具体请参见表 IQueryUserCouponsResult。| |参数约束以及描述：优惠券记录。具体请参见表 IQueryUserCouponsResult。|
	UserCoupons    *[]IQueryUserPartnerCouponsResultV2 `json:"user_coupons,omitempty"`
	HttpStatusCode int                                 `json:"-"`
}

func (o ListIssuedPartnerCouponsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIssuedPartnerCouponsResponse struct{}"
	}

	return strings.Join([]string{"ListIssuedPartnerCouponsResponse", string(data)}, " ")
}
