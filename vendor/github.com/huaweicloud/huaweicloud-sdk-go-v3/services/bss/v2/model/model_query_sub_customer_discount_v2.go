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

type QuerySubCustomerDiscountV2 struct {
	// |参数名称：折扣ID，唯一的表示一条折扣信息。| |参数约束及描述：折扣ID，唯一的表示一条折扣信息。|
	DiscountId *string `json:"discount_id,omitempty"`
	// |参数名称：折扣率，精确到4位小数。如果折扣率是22%，则折扣率写成0.22。| |参数的约束及描述：折扣率，精确到4位小数。如果折扣率是22%，则折扣率写成0.22。|
	Discount float32 `json:"discount,omitempty"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”|
	ExpireTime *string `json:"expire_time,omitempty"`
}

func (o QuerySubCustomerDiscountV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuerySubCustomerDiscountV2 struct{}"
	}

	return strings.Join([]string{"QuerySubCustomerDiscountV2", string(data)}, " ")
}
