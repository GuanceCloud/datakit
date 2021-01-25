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

type CouponInfoV2 struct {
	// |参数名称：优惠券实例ID。| |参数约束及描述：优惠券实例ID。|
	CouponId *string `json:"coupon_id,omitempty"`
	// |参数名称：优惠券编码。| |参数约束及描述：优惠券编码。|
	CouponCode *string `json:"coupon_code,omitempty"`
	// |参数名称：优惠券状态：2：待使用。| |参数的约束及描述：优惠券状态：2：待使用。|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：优惠券类型：301：代金券；302：现金券。| |参数的约束及描述：优惠券类型：301：代金券；302：现金券。|
	CouponType *int32 `json:"coupon_type,omitempty"`
	// |参数名称：面额单位：1：元。| |参数的约束及描述：面额单位：1：元。|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：优惠券金额。| |参数的约束及描述：优惠券金额。|
	FaceValue *float64 `json:"face_value,omitempty"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-dTHH:mm:ssZ，如2019-05-06T08:05:01Z。| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-dTHH:mm:ssZ，如2019-05-06T08:05:01Z。|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：促销计划名称。| |参数约束及描述：促销计划名称。|
	PlanName *string `json:"plan_name,omitempty"`
	// |参数名称：促销计划描述。| |参数约束及描述：促销计划描述。|
	PlanDesc *string `json:"plan_desc,omitempty"`
	// |参数名称：优惠券限制。具体请参见表 LimitInfo。| |参数约束以及描述：优惠券限制。具体请参见表 LimitInfo。|
	UseLimits *[]LimitInfoV2 `json:"use_limits,omitempty"`
	// |参数名称：激活时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。| |参数约束及描述：激活时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。|
	ActiveTime *string `json:"active_time,omitempty"`
	// |参数名称：上一次使用时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。| |参数约束及描述：上一次使用时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。|
	LastUsedTime *string `json:"last_used_time,omitempty"`
	// |参数名称：创建时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。| |参数约束及描述：创建时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如2019-05-06T08:05:01Z。|
	CreateTime *string `json:"create_time,omitempty"`
	// |参数名称：优惠券版本。1：老版本（原本分为三种：代金券/折扣券/奖金券）；2：新版本（只有代金券）。| |参数的约束及描述：优惠券版本。1：老版本（原本分为三种：代金券/折扣券/奖金券）；2：新版本（只有代金券）。|
	CouponVersion *int32 `json:"coupon_version,omitempty"`
	// |参数名称：余额。| |参数约束及描述： 余额。|
	Balance *float64 `json:"balance,omitempty"`
	// |参数名称：使用优惠券的订单ID，表示正在有另外一张订单正在使用这个优惠券。正在锁定的时候，只有锁定优惠券的订单才能使用这个优惠券，其他订单不能使用该优惠券。| |参数约束及描述：使用优惠券的订单ID，表示正在有另外一张订单正在使用这个优惠券。正在锁定的时候，只有锁定优惠券的订单才能使用这个优惠券，其他订单不能使用该优惠券。|
	UsedByOrderId *string `json:"used_by_order_id,omitempty"`
	// |参数名称：优惠券用途。| |参数约束及描述：优惠券用途。|
	CouponUsage *string `json:"coupon_usage,omitempty"`
}

func (o CouponInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CouponInfoV2 struct{}"
	}

	return strings.Join([]string{"CouponInfoV2", string(data)}, " ")
}
