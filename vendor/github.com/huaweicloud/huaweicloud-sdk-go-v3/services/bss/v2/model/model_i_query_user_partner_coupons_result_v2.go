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

type IQueryUserPartnerCouponsResultV2 struct {
	// |参数名称：优惠券实例ID。| |参数约束及描述：优惠券实例ID。|
	CouponId *string `json:"coupon_id,omitempty"`
	// |参数名称：优惠券状态：1：未激活；2：待使用；3：已使用；4：已过期。| |参数的约束及描述：优惠券状态：1：未激活；2：待使用；3：已使用；4：已过期。|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：客户ID| |参数约束及描述：客户ID|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：优惠券类型：1：代金券；4：现金券。| |参数的约束及描述：优惠券类型：1：代金券；4：现金券。|
	CouponType *int32 `json:"coupon_type,omitempty"`
	// |参数名称：度量单位。1：元| |参数的约束及描述：度量单位。1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：优惠券金额。| |参数的约束及描述：优惠券金额。|
	FaceValue float32 `json:"face_value,omitempty"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：订单ID。| |参数约束及描述：订单ID。|
	OrderId *string `json:"order_id,omitempty"`
	// |参数名称：促销计划ID。| |参数约束及描述：促销计划ID。|
	PromotionPlanId *string `json:"promotion_plan_id,omitempty"`
	// |参数名称：促销计划名称。| |参数约束及描述：促销计划名称。|
	PromotionPlanName *string `json:"promotion_plan_name,omitempty"`
	// |参数名称：促销计划描述。| |参数约束及描述：促销计划描述。|
	PromotionPlanDesc *string `json:"promotion_plan_desc,omitempty"`
	// |参数名称：介质类型。| |参数的约束及描述：介质类型。|
	MediaType *int32 `json:"media_type,omitempty"`
	// |参数名称：获取方式：1：线上领取；2：线上兑换；3：线上发放；4：线下获取；5：事件赠送。| |参数的约束及描述：获取方式：1：线上领取；2：线上兑换；3：线上发放；4：线下获取；5：事件赠送。|
	FetchMethod *int32 `json:"fetch_method,omitempty"`
	// |参数名称：优惠券使用限制。具体请参见表 ICouponUseLimitInfo。| |参数约束以及描述：优惠券使用限制。具体请参见表 ICouponUseLimitInfo。|
	UseLimits *[]ICouponUseLimitInfoV2 `json:"use_limits,omitempty"`
	// |参数名称：激活时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：激活时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	ActiveTime *string `json:"active_time,omitempty"`
	// |参数名称：使用时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：使用时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	LastUsedTime *string `json:"last_used_time,omitempty"`
	// |参数名称：促销ID。| |参数约束及描述：促销ID。|
	PromotionId *string `json:"promotion_id,omitempty"`
	// |参数名称：创建时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：创建时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ|
	CreateTime *string `json:"create_time,omitempty"`
	// |参数名称：余额。如果为老版本优惠券，该值为空| |参数的约束及描述：余额。如果为老版本优惠券，该值为空|
	Balance float32 `json:"balance,omitempty"`
	// |参数名称：锁定优惠券的订单ID。如果为老版本优惠券，该值为空。| |参数约束及描述：锁定优惠券的订单ID。如果为老版本优惠券，该值为空。|
	LockOrderId *string `json:"lock_order_id,omitempty"`
	// |参数名称：优惠券是否冻结：0：否1：是可用优惠券接口返回时不包括冻结状态的优惠券。| |参数约束及描述：优惠券是否冻结：0：否1：是可用优惠券接口返回时不包括冻结状态的优惠券。|
	IsFrozen *string `json:"is_frozen,omitempty"`
}

func (o IQueryUserPartnerCouponsResultV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IQueryUserPartnerCouponsResultV2 struct{}"
	}

	return strings.Join([]string{"IQueryUserPartnerCouponsResultV2", string(data)}, " ")
}
