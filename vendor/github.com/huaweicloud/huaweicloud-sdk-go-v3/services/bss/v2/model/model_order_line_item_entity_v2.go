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

type OrderLineItemEntityV2 struct {
	// |参数名称：订单项Id。| |参数约束及描述：订单项Id。|
	OrderLineItemId *string `json:"order_line_item_id,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：产品ID。| |参数约束及描述：产品ID。|
	ProductId *string `json:"product_id,omitempty"`
	// |参数名称：产品规格描述。| |参数约束及描述：产品规格描述。|
	ProductSpecDesc *string `json:"product_spec_desc,omitempty"`
	// |参数名称：周期类型。0：天；1：周；2：月；3：年；4：小时；5：一次性；6：按需（预留）；7：按用量报表使用（预留）。| |参数的约束及描述：周期类型。0：天；1：周；2：月；3：年；4：小时；5：一次性；6：按需（预留）；7：按用量报表使用（预留）。|
	PeriodType *int32 `json:"period_type,omitempty"`
	// |参数名称：周期数量。| |参数的约束及描述：周期数量。|
	PeriodNum *int32 `json:"period_num,omitempty"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：订购数量。| |参数的约束及描述：订购数量。|
	SubscriptionNum *int32 `json:"subscription_num,omitempty"`
	// |参数名称：订单优惠后金额（实付价格，不含券不含卡）。| |参数的约束及描述：订单优惠后金额（实付价格，不含券不含卡）。|
	AmountAfterDiscount *float64 `json:"amount_after_discount,omitempty"`
	// |参数名称：订单金额（官网价）。退订订单中，该金额等于currencyAfterDiscount。| |参数的约束及描述：订单金额（官网价）。退订订单中，该金额等于currencyAfterDiscount。|
	OfficialAmount *float64            `json:"official_amount,omitempty"`
	AmountInfo     *AmountInfomationV2 `json:"amount_info,omitempty"`
	// |参数名称：货币编码。| |参数约束及描述：货币编码。如CNY|
	Currency *string `json:"currency,omitempty"`
	// |参数名称：产品目录编码。| |参数约束及描述：产品目录编码。|
	CategoryCode *string `json:"category_code,omitempty"`
	// |参数名称：产品归属的云服务类型编码。| |参数约束及描述：产品归属的云服务类型编码。|
	ProductOwnerService *string `json:"product_owner_service,omitempty"`
	// |参数名称：商务归属的资源类型编码。| |参数约束及描述：商务归属的资源类型编码。|
	CommercialResource *string `json:"commercial_resource,omitempty"`
}

func (o OrderLineItemEntityV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OrderLineItemEntityV2 struct{}"
	}

	return strings.Join([]string{"OrderLineItemEntityV2", string(data)}, " ")
}
