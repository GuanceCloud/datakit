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

type MonthlyBillRes struct {
	// |参数名称：消费月份| |参数的约束及描述：格式为YYYY-MM|
	Cycle *string `json:"cycle,omitempty"`
	// |参数名称：账单类型| |参数的约束及描述：该参数非必填，1：消费-新购；2：消费-续订；3：消费-变更；4：退款-退订；5：消费-使用；8：消费-自动续订；9：调账-补偿；12：消费-按时计费；13：消费-退订手续费； 15消费-税金；14：消费-服务支持计划月末扣费；16：调账-扣费 100：退款-退订税金 101：调账-补偿税金 102：调账-扣费税金|
	BillType *int32 `json:"bill_type,omitempty"`
	// |参数名称：消费的客户账号ID。| |参数约束及描述：如果是普通客户或者企业子客户查询消费记录，只能查询到客户自己的消费记录，且此处显示的是客户自己的客户ID; 如果是企业主查询消费记录，可以查询到企业主以及企业子客户的消费记录，此处为消费的实际客户ID。如果是企业主自己的消费记录，则为企业主ID；如果是某个企业子客户的消费记录，则此处为企业子账号ID。|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：云服务区编码| |参数的约束及描述：该参数非必填，例如：“cn-north-1”。|
	Region *string `json:"region,omitempty"`
	// |参数名称：云服务区名称| |参数的约束及描述：云服务区名称|
	RegionName *string `json:"region_name,omitempty"`
	// |参数名称：云服务类型编码| |参数的约束及描述：该参数非必填,，例如ECS的云服务类型编码为“hws.service.type.ec2”|
	CloudServiceType *string `json:"cloud_service_type,omitempty"`
	// |参数名称：资源类型编码| |参数的约束及描述：该参数非必填，例如ECS的VM为“hws.resource.type.vm”。|
	ResourceTypeCode *string `json:"resource_Type_code,omitempty"`
	// |参数名称：资源实例ID| |参数的约束及描述：该参数非必填|
	ResInstanceId *string `json:"res_instance_id,omitempty"`
	// |参数名称：资源名称| |参数的约束及描述：客户在创建资源的时候，可以输入资源名称，有些资源也可以在管理资源时，修改资源名称|
	ResourceName *string `json:"resource_name,omitempty"`
	// |参数名称：资源标签| |参数的约束及描述：客户在创建资源的时候，可以输入资源名称，有些资源也可以在管理资源时，修改资源名称|
	ResourceTag *string `json:"resource_tag,omitempty"`
	// |参数名称：SKU编码| |参数的约束及描述：SKU（Stock Keeping Unit，库存量单元）编码，产品下的SKU分类属性|
	SkuCode *string `json:"sku_code,omitempty"`
	// |参数名称：企业项目ID| |参数的约束及描述：该参数非必填|
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// |参数名称：企业项目名称| |参数的约束及描述：该参数非必填|
	EnterpriseProjectName *string `json:"enterprise_project_name,omitempty"`
	// |参数名称：计费模式| |参数的约束及描述：1 : 包年/包月；3: 按需。10: 预留实例|
	ChargeMode *int32 `json:"charge_mode,omitempty"`
	// |参数名称：客户购买云服务类型的消费金额| |参数的约束及描述：该参数非必填，包含代金券，大陆站还包含现金券，大陆站精确到小数点后8位，国际站精确到小数点后2位。|
	ConsumeAmount float32 `json:"consume_amount,omitempty"`
	// |参数名称：现金支付金额| |参数的约束及描述：该参数非必填|
	CashAmount float32 `json:"cash_amount,omitempty"`
	// |参数名称：信用额度支付金额| |参数的约束及描述：该参数非必填|
	CreditAmount float32 `json:"credit_amount,omitempty"`
	// |参数名称：代金券支付金额| |参数的约束及描述：该参数非必填。|
	CouponAmount float32 `json:"coupon_amount,omitempty"`
	// |参数名称：现金券支付金额| |参数的约束及描述：该参数非必填。|
	FlexipurchaseCouponAmount float32 `json:"flexipurchase_coupon_amount,omitempty"`
	// |参数名称：储值卡支付金额| |参数的约束及描述：该参数非必填。|
	StoredCardAmount float32 `json:"stored_card_amount,omitempty"`
	// |参数名称：奖励金支付金额（用于现网未清干净的奖励金）| |参数的约束及描述：该参数非必填。|
	BonusAmount float32 `json:"bonus_amount,omitempty"`
	// |参数名称：欠费金额| |参数的约束及描述：该参数非必填。|
	DebtAmount float32 `json:"debt_amount,omitempty"`
	// |参数名称：欠费核销金额| |参数的约束及描述：该参数非必填。|
	AdjustmentAmount float32 `json:"adjustment_amount,omitempty"`
	// |参数名称：官网价| |参数的约束及描述：该参数非必填。|
	OfficialAmount float32 `json:"official_amount,omitempty"`
	// |参数名称：对应官网价折扣金额| |参数的约束及描述：该参数非必填。|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：金额单位。1:元| |参数的约束及描述：该参数非必填|
	MeasureId *int32 `json:"measure_id,omitempty"`
}

func (o MonthlyBillRes) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MonthlyBillRes struct{}"
	}

	return strings.Join([]string{"MonthlyBillRes", string(data)}, " ")
}
