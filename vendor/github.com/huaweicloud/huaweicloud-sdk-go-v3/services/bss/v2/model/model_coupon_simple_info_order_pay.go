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

type CouponSimpleInfoOrderPay struct {
	// |参数名称：优惠券ID，同种类型的优惠券，列表前面会优先使用| |参数约束及描述：优惠券ID，同种类型的优惠券，列表前面会优先使用|
	Id string `json:"id"`
	// |参数名称：折扣类型：取值为300-折扣卷 301-促销代金券302-促销现金券303-促销储值卡| |参数的约束及描述：折扣类型：取值为300-折扣卷 301-促销代金券302-促销现金券303-促销储值卡|
	Type int32 `json:"type"`
}

func (o CouponSimpleInfoOrderPay) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CouponSimpleInfoOrderPay struct{}"
	}

	return strings.Join([]string{"CouponSimpleInfoOrderPay", string(data)}, " ")
}
