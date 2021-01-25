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

type BillSumInfoV2 struct {
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型。|
	CloudServiceType *string `json:"cloud_service_type,omitempty"`
	// |参数名称：费用类型。0：消费；1：退订；2：华为核销。| |参数约束及描述：费用类型。0：消费；1：退订；2：华为核销。|
	BillType *string `json:"bill_type,omitempty"`
	// |参数名称：消费类型。1：包周期；3: 按需。| |参数约束及描述：消费类型。1：包周期；3: 按需。|
	ChargeMode *string `json:"charge_mode,omitempty"`
	// |参数名称：消费的金额，即从客户账户实际扣除的金额。对于billType=1或者2的账单，该金额为负值。| |参数的约束及描述：消费的金额，即从客户账户实际扣除的金额。对于billType=1或者2的账单，该金额为负值。|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：欠费金额，指从客户账户扣费的时候，客户账户金额不足，欠费的金额，华为核销或者退订的时候没有该字段。| |参数的约束及描述：欠费金额，指从客户账户扣费的时候，客户账户金额不足，欠费的金额，华为核销或者退订的时候没有该字段。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：核销欠款，华为核销或者退订的时候没有该字段。| |参数的约束及描述：核销欠款，华为核销或者退订的时候没有该字段。|
	AdjustmentAmount float32 `json:"adjustment_amount,omitempty"`
	// |参数名称：折扣金额，华为核销或者退订的时候没有该字段。| |参数的约束及描述：折扣金额，华为核销或者退订的时候没有该字段。|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：金额单位。1：元；2：角；3：分| |参数的约束及描述：金额单位。1：元；2：角；3：分|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：按不同账户消费类型和付费方式区分的支付总金额。具体请参见表 BalanceTypeDeductSum。| |参数约束以及描述：按不同账户消费类型和付费方式区分的支付总金额。具体请参见表 BalanceTypeDeductSum。|
	AccountDetails *[]BalanceTypeDeductSumV2 `json:"account_details,omitempty"`
	// |参数名称：资源类型编码| |参数约束及描述：资源类型编码|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
}

func (o BillSumInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BillSumInfoV2 struct{}"
	}

	return strings.Join([]string{"BillSumInfoV2", string(data)}, " ")
}
