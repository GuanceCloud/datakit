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

type ICouponUseLimitInfoV2 struct {
	// |参数名称：使用限制ID，主键。| |参数约束及描述：使用限制ID，主键。|
	UseLimitiInfoId *string `json:"use_limiti_info_id,omitempty"`
	// |参数名称：折扣限制，key的取值请参考表ICouponUseLimitInfo的limit_key要求。| |参数约束及描述：折扣限制，key的取值请参考表 ICouponUseLimitInfo的limit_key要求。|
	LimitKey *string `json:"limit_key,omitempty"`
	// |参数名称：value1。| |参数约束及描述：value1。|
	Value1 *string `json:"value1,omitempty"`
	// |参数名称：value2。| |参数约束及描述：value2。|
	Value2 *string `json:"value2,omitempty"`
	// |参数名称：value单位。| |参数约束及描述：value单位。|
	ValueUnit *string `json:"value_unit,omitempty"`
	// |参数名称：限制类型。| |参数约束及描述：限制类型。|
	LimitType *string `json:"limit_type,omitempty"`
	// |参数名称：促销计划ID。| |参数约束及描述：促销计划ID。|
	PromotionPlanId *string `json:"promotion_plan_id,omitempty"`
}

func (o ICouponUseLimitInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ICouponUseLimitInfoV2 struct{}"
	}

	return strings.Join([]string{"ICouponUseLimitInfoV2", string(data)}, " ")
}
