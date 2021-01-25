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

type MonthlyBillRecord struct {
	// |参数名称：账期，格式固定为YYYY-MM。| |参数约束及描述：账期，格式固定为YYYY-MM。|
	BillCycle *string `json:"bill_cycle,omitempty"`
	// |参数名称：消费的客户账号ID。如果是普通客户或者企业子客户查询消费记录，只能查询到客户自己的消费记录，且此处显示的是客户自己的客户ID。如果是企业主查询消费记录，可以查询到企业主以及企业子客户的消费记录，此处为消费的实际客户ID。如果是企业主自己的消费记录，则为企业主ID；如果是某个企业子客户的消费记录，则此处为企业子账号ID。| |参数约束及描述：消费的客户账号ID。如果是普通客户或者企业子客户查询消费记录，只能查询到客户自己的消费记录，且此处显示的是客户自己的客户ID。如果是企业主查询消费记录，可以查询到企业主以及企业子客户的消费记录，此处为消费的实际客户ID。如果是企业主自己的消费记录，则为企业主ID；如果是某个企业子客户的消费记录，则此处为企业子账号ID。|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。| |参数约束及描述：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型| |参数约束及描述：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型|
	RegionCode *string `json:"region_code,omitempty"`
	// |参数名称：企业项目标识| |参数约束及描述：企业项目标识|
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// |参数名称：企业项目名称| |参数约束及描述：企业项目名称|
	EnterpriseProjectName *string `json:"enterprise_project_name,omitempty"`
	// |参数名称：计费模式1、包周期；3、按需；10、预留实例| |参数的约束及描述：计费模式1、包周期；3、按需；10、预留实例|
	ChargingMode *int32 `json:"charging_mode,omitempty"`
	// |参数名称：| |参数名称：消费时间，包周期和预留实例订购场景下为订单支付时间，按需场景下为话单生失效时间||参数约束及描述：| |参数名称：消费时间，包周期和预留实例订购场景下为订单支付时间，按需场景下为话单生失效时间|
	ConsumeTime *string `json:"consume_time,omitempty"`
	// |参数名称：| |参数名称：交易时间| |参数约束及描述：交易时间，某条消费记录对应的扣费时间||参数约束及描述：| |参数名称：交易时间| |参数约束及描述：交易时间，某条消费记录对应的扣费时间|
	TradeTime *string `json:"trade_time,omitempty"`
	// |参数名称：服务商1：华为云2：云市场| |参数的约束及描述：服务商1：华为云2：云市场|
	ProviderType *int32 `json:"provider_type,omitempty"`
	// |参数名称：订单ID 或 交易ID，扣费维度的唯一标识| |参数约束及描述：订单ID 或 交易ID，扣费维度的唯一标识|
	TradeId *string `json:"trade_id,omitempty"`
	// |参数名称：账单类型1：消费-新购2：消费-续订3：消费-变更4：退款-退订5：消费-使用8：消费-自动续订9：调账-补偿12：消费-按时计费13：消费-退订手续费14：消费-服务支持计划月末扣费16：调账-扣费| |参数的约束及描述：账单类型1：消费-新购2：消费-续订3：消费-变更4：退款-退订5：消费-使用8：消费-自动续订9：调账-补偿12：消费-按时计费13：消费-退订手续费14：消费-服务支持计划月末扣费16：调账-扣费|
	BillType *int32 `json:"bill_type,omitempty"`
	// |参数名称：支付状态1：已支付2：未结清3：未结算| |参数的约束及描述：支付状态1：已支付2：未结清3：未结算|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：官网价。单位为元说明：official_amount等于official_discount_amount和erase_amount和consume_amount的和。| |参数的约束及描述：官网价。单位为元说明：official_amount等于official_discount_amount和erase_amount和consume_amount的和。|
	OfficialAmount float32 `json:"official_amount,omitempty"`
	// |参数名称：折扣金额。单位为元| |参数的约束及描述：折扣金额。单位为元|
	OfficialDiscountAmount float32 `json:"official_discount_amount,omitempty"`
	// |参数名称：抹零金额。单位为元| |参数的约束及描述：抹零金额。单位为元|
	EraseAmount float32 `json:"erase_amount,omitempty"`
	// |参数名称：应付金额，包括现金券和储值卡和代金券金额。单位为元说明：（1）consume_amount的值包含cash_amount，credit_amount，coupon_amount，flexipurchase_coupon_amount，stored_card_amount，bonus_amount，debt_amount，writeoff_amount的和| |参数的约束及描述：应付金额，包括现金券和储值卡和代金券金额。单位为元说明：（1）consume_amount的值包含cash_amount，credit_amount，coupon_amount，flexipurchase_coupon_amount，stored_card_amount，bonus_amount，debt_amount，writeoff_amount的和|
	ConsumeAmount float32 `json:"consume_amount,omitempty"`
	// |参数名称：现金支付金额。单位为元| |参数的约束及描述：现金支付金额。单位为元|
	CashAmount float32 `json:"cash_amount,omitempty"`
	// |参数名称：信用额度支付金额。单位为元| |参数的约束及描述：信用额度支付金额。单位为元|
	CreditAmount float32 `json:"credit_amount,omitempty"`
	// |参数名称：代金券支付金额。单位为元| |参数的约束及描述：代金券支付金额。单位为元|
	CouponAmount float32 `json:"coupon_amount,omitempty"`
	// |参数名称：现金券支付金额。单位为元| |参数的约束及描述：现金券支付金额。单位为元|
	FlexipurchaseCouponAmount float32 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：储值卡支付金额。单位为元| |参数的约束及描述：储值卡支付金额。单位为元|
	StoredValueCardAmount float32 `json:"stored_value_card_amount,omitempty"`
	// |参数名称：奖励金支付金额（奖励金已经下市，用于现网客户未使用完的奖励金）。单位为元| |参数的约束及描述：奖励金支付金额（奖励金已经下市，用于现网客户未使用完的奖励金）。单位为元|
	BonusAmount float32 `json:"bonus_amount,omitempty"`
	// |参数名称：欠费金额。单位为元| |参数的约束及描述：欠费金额。单位为元|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：欠费核销金额。单位为元| |参数的约束及描述：欠费核销金额。单位为元|
	WriteoffAmount float32 `json:"writeoff_amount,omitempty"`
	// |参数名称：云服务区名称| |参数的约束及描述：云服务区名称|
	RegionName *string `json:"region_name,omitempty"`
}

func (o MonthlyBillRecord) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MonthlyBillRecord struct{}"
	}

	return strings.Join([]string{"MonthlyBillRecord", string(data)}, " ")
}
