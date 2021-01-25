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

type SetSubCustomerDiscountV2 struct {
	// |参数名称：客户ID| |参数约束及描述：客户ID|
	CustomerId string `json:"customer_id"`
	// |参数名称：折扣率，最高精确到4位小数。折扣范围：0.8~1。如果折扣率是85%，则折扣率写成0.85。注意：折扣为1表示不打折，相当于删除伙伴折扣。| |参数的约束及描述：折扣率，最高精确到4位小数。折扣范围：0.8~1。如果折扣率是85%，则折扣率写成0.85。注意：折扣为1表示不打折，相当于删除伙伴折扣。|
	Discount float32 `json:"discount"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	ExpireTime *string `json:"expire_time,omitempty"`
}

func (o SetSubCustomerDiscountV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetSubCustomerDiscountV2 struct{}"
	}

	return strings.Join([]string{"SetSubCustomerDiscountV2", string(data)}, " ")
}
