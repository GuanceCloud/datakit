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

type SubCustomerResFeeRecordV2 struct {
	// |参数名称：费用对应的资源使用的开始时间，按需有效，包周期该字段保留。| |参数约束及描述：费用对应的资源使用的开始时间，按需有效，包周期该字段保留。|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：费用对应的资源使用的结束时间，按需有效，包周期该字段保留。| |参数约束及描述：费用对应的资源使用的结束时间，按需有效，包周期该字段保留。|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：产品ID。| |参数约束及描述：产品ID。|
	ProductId *string `json:"product_id,omitempty"`
	// |参数名称：产品名称。| |参数约束及描述：产品名称。|
	ProductName *string `json:"product_name,omitempty"`
	// |参数名称：订单ID，包周期资源使用记录才有该字段，按需资源为空。| |参数约束及描述：订单ID，包周期资源使用记录才有该字段，按需资源为空。|
	OrderId *string `json:"order_id,omitempty"`
	// |参数名称：消费金额，包括现金券和代金券金额，精确到小数点后2位。| |参数约束及描述： 消费金额，包括现金券和代金券金额，精确到小数点后2位。|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：金额单位：1：元；2：角；3：分。| |参数的约束及描述：金额单位：1：元；2：角；3：分。|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：使用量类型| |参数约束及描述：使用量类型|
	UsageType *string `json:"usage_type,omitempty"`
	// |参数名称：使用量。| |参数约束及描述： 使用量。|
	Usage float32 `json:"usage,omitempty"`
	// |参数名称：使用量度量单位| |参数的约束及描述：使用量度量单位|
	UsageMeasureId *int32 `json:"usage_measure_id,omitempty"`
	// |参数名称：套餐内使用量。| |参数约束及描述： 套餐内使用量。|
	FreeResourceUsage float32 `json:"free_resource_usage,omitempty"`
	// |参数名称：套餐内使用量单位，具体枚举参考：usage_measure_id| |参数的约束及描述：套餐内使用量单位，具体枚举参考：usage_measure_id|
	FreeResourceMeasureId *int32 `json:"free_resource_measure_id,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	CloudServiceType *string `json:"cloud_service_type,omitempty"`
	// |参数名称：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。| |参数约束及描述：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。|
	Region *string `json:"region,omitempty"`
	// |参数名称：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。| |参数约束及描述：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。|
	ResourceType *string `json:"resource_type,omitempty"`
	// |参数名称：1 : 包周期；3: 按需。10: 预留实例。| |参数约束及描述：1 : 包周期；3: 按需。10: 预留实例。|
	ChargeMode *string `json:"charge_mode,omitempty"`
	// |参数名称：资源标签。| |参数约束及描述：资源标签。|
	ResourceTag *string `json:"resource_tag,omitempty"`
	// |参数名称：资源名称。| |参数约束及描述：资源名称。|
	ResourceName *string `json:"resource_name,omitempty"`
	// |参数名称：资源ID。| |参数约束及描述：资源ID。|
	ResourceId *string `json:"resource_id,omitempty"`
	// |参数名称：账单类型。1：消费-新购2：消费-续订3：消费-变更4：退款-退订5：消费-使用8：消费-自动续订9：调账-补偿12：消费-按时计费13：消费-退订手续费14：消费-服务支持计划月末扣费16：调账-扣费| |参数的约束及描述：账单类型。1：消费-新购2：消费-续订3：消费-变更4：退款-退订5：消费-使用8：消费-自动续订9：调账-补偿12：消费-按时计费13：消费-退订手续费14：消费-服务支持计划月末扣费16：调账-扣费|
	BillType *int32 `json:"bill_type,omitempty"`
	// |参数名称：周期类型：19：年；20：月；24：天；25：小时；5：分钟；6：秒。| |参数约束及描述：周期类型：19：年；20：月；24：天；25：小时；5：分钟；6：秒。|
	PeriodType *string `json:"period_type,omitempty"`
	// |参数名称：产品规格描述。| |参数约束及描述：产品规格描述，举例为：普通IO|100.0GB。|
	ProductSpecDesc *string `json:"product_spec_desc,omitempty"`
	// |参数名称：预留实例使用量。| |参数约束及描述： 预留实例使用量。|
	RiUsage float32 `json:"ri_usage,omitempty"`
	// |参数名称：预留实例使用量单位。| |参数的约束及描述：预留实例使用量单位。|
	RiUsageMeasureId *int32 `json:"ri_usage_measure_id,omitempty"`
	// |参数名称：官网价。| |参数约束及描述： 官网价。|
	OfficialAmount float32 `json:"official_amount,omitempty"`
	// |参数名称：折扣金额| |参数约束及描述： 折扣金额|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：现金支付金额| |参数约束及描述： 现金支付金额|
	CashAmount float32 `json:"cash_amount,omitempty"`
	// |参数名称：信用额度支付金额。| |参数约束及描述： 信用额度支付金额。|
	CreditAmount float32 `json:"credit_amount,omitempty"`
	// |参数名称：代金券支付金额。| |参数约束及描述： 代金券支付金额。|
	CouponAmount float32 `json:"coupon_amount,omitempty"`
	// |参数名称：现金券支付金额。| |参数约束及描述： 现金券支付金额。|
	FlexipurchaseCouponAmount float32 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：储值卡支付金额。| |参数约束及描述： 储值卡支付金额。|
	StoredCardAmount float32 `json:"stored_card_amount,omitempty"`
	// |参数名称：奖励金支付金额（用于现网未清干净的奖励金）。| |参数约束及描述： 奖励金支付金额（用于现网未清干净的奖励金）。|
	BonusAmount float32 `json:"bonus_amount,omitempty"`
	// |参数名称：欠费金额。| |参数约束及描述： 欠费金额。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：欠费核销金额。| |参数约束及描述： 欠费核销金额。|
	AdjustmentAmount float32 `json:"adjustment_amount,omitempty"`
	// |参数名称：线性大小| |参数约束及描述： 线性大小|
	SpecSize float32 `json:"spec_size,omitempty"`
	// |参数名称：线性大小单位| |参数的约束及描述：线性大小单位|
	SpecSizeMeasureId *int32 `json:"spec_size_measure_id,omitempty"`
	// |参数名称：云服务区名称| |参数的约束及描述：云服务区名称|
	RegionName *string `json:"region_name,omitempty"`
}

func (o SubCustomerResFeeRecordV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SubCustomerResFeeRecordV2 struct{}"
	}

	return strings.Join([]string{"SubCustomerResFeeRecordV2", string(data)}, " ")
}
