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

type BillSumRecordInfoV2 struct {
	// |参数名称：账期，格式为YYYY-MM| |参数约束及描述：账期，格式为YYYY-MM|
	BillCycle *string `json:"bill_cycle,omitempty"`
	// |参数名称：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。| |参数约束及描述：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：计费模式1：包周期；3：按需；10：预留实例| |参数约束及描述：计费模式1：包周期；3：按需；10：预留实例|
	ChargingMode *int32 `json:"charging_mode,omitempty"`
	// |参数名称：官网价| |参数的约束及描述：官网价|
	OfficialAmount float32 `json:"official_amount,omitempty"`
	// |参数名称：折扣金额| |参数的约束及描述：折扣金额|
	OfficialDiscountAmount float32 `json:"official_discount_amount,omitempty"`
	// |参数名称：抹零金额| |参数的约束及描述：抹零金额|
	TruncatedAmount float32 `json:"truncated_amount,omitempty"`
	// |参数名称：应付金额，应付金额 = 官网价-折扣金额-抹零金额| |参数的约束及描述：应付金额，应付金额 = 官网价-折扣金额-抹零金额|
	ConsumeAmount float32 `json:"consume_amount,omitempty"`
	// |参数名称：代金券金额。| |参数的约束及描述：代金券金额。|
	CouponAmount float32 `json:"coupon_amount,omitempty"`
	// |参数名称：现金券金额，预留。| |参数的约束及描述：现金券金额，预留。|
	FlexipurchaseCouponAmount float32 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：储值卡金额，预留。| |参数的约束及描述：储值卡金额，预留。|
	StoredValueCardAmount float32 `json:"stored_value_card_amount,omitempty"`
	// |参数名称：欠费金额，即从客户账户扣费的时候，客户账户金额不足，欠费的金额。| |参数的约束及描述：欠费金额，即从客户账户扣费的时候，客户账户金额不足，欠费的金额。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：欠费核销金额| |参数的约束及描述：欠费核销金额|
	WriteoffAmount float32 `json:"writeoff_amount,omitempty"`
	// |参数名称：现金账户金额。| |参数的约束及描述：现金账户金额。|
	CashAmount float32 `json:"cash_amount,omitempty"`
	// |参数名称：信用账户金额。| |参数的约束及描述：信用账户金额。|
	CreditAmount float32 `json:"credit_amount,omitempty"`
	// |参数名称：金额单位。1：元| |参数的约束及描述：金额单位。1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：消费类型：1：消费2：退款3：调账| |参数的约束及描述：消费类型：1：消费2：退款3：调账|
	BillType *int32 `json:"bill_type,omitempty"`
	// |参数名称：消费的客户账号ID。| |参数约束及描述：如果是查询自己，这个地方是自身的ID; 如果是查询某个企业子客户，这个地方是企业子客户ID如果是查询以及下面的所有子客户，这个地方是消费的实际客户ID; 如果是企业主自身消费，为企业主ID，如果这条消费记录是某个企业子客户的消费，这个地方的ID是企业子账号ID。|
	CustomerId *string `json:"customer_id,omitempty"`
}

func (o BillSumRecordInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BillSumRecordInfoV2 struct{}"
	}

	return strings.Join([]string{"BillSumRecordInfoV2", string(data)}, " ")
}
