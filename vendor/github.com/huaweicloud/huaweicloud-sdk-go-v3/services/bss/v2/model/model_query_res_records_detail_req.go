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

type QueryResRecordsDetailReq struct {
	// |参数名称：消费月份| |参数的约束及描述：该参数必填，最大长度：8，比如2018-12|
	Cycle string `json:"cycle"`
	// |参数名称：云服务类型编码| |参数的约束及描述：该参数非必填，最大长度：64，且只允许字符串，例如ECS的云服务类型编码为“hws.service.type.ec2”|
	CloudServiceType *string `json:"cloud_service_type,omitempty"`
	// |参数名称：资源类型编码| |参数的约束及描述：该参数非必填，最大长度：64，且只允许字符串，例如ECS的VM为“hws.resource.type.vm”|
	ResourceType *string `json:"resource_type,omitempty"`
	// |参数名称：云服务区编码| |参数的约束及描述：该参数非必填，最大长度：64，且只允许字符串，例如：“cn-north-1”|
	Region *string `json:"region,omitempty"`
	// |参数名称：资源实例ID| |参数的约束及描述：该参数非必填，最大长度：64，且只允字符串|
	ResInstanceId *string `json:"res_instance_id,omitempty"`
	// |参数名称：支付方式| |参数的约束及描述：该参数非必填，且只允许整数,1 : 包周期；3: 按需。10: 预留实例|
	ChargeMode *int32 `json:"charge_mode,omitempty"`
	// |参数名称：账单类型| |参数的约束及描述：该参数非必填，且只允许整数,1：消费-新购；2：消费-续订；3：消费-变更；4：退款-退订；5：消费-使用；8：消费-自动续订；9：调账-补偿；12：消费-按时计费；13：消费-退订手续费；14：消费-服务支持计划月末扣费； 15消费-税金；16：调账-扣费; 17：消费-保底差额 100：退款-退订税金 101：调账-补偿税金 102：调账-扣费税金|
	BillType *int32 `json:"bill_type,omitempty"`
	// |参数名称：企业项目ID| |参数的约束及描述：该参数非必，最大长度：64，且只允许字符串|
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// |参数名称：返回是否包含应付金额为0的记录| |参数的约束及描述：该参数非必填，且只允许布尔型，true: 包含；false: 不包含|
	IncludeZeroRecord *bool `json:"include_zero_record,omitempty"`
	// |参数名称：偏移量| |参数的约束及描述：该参数非必填，且只允许数字，默认为1|
	Offset *int32 `json:"offset,omitempty"`
	// |参数名称：页面大小| |参数的约束及描述：该参数非必填，且只允许1-100的数字，默认10|
	Limit *int32 `json:"limit,omitempty"`
	// |参数名称：查询方式。oneself：自身sub_customer: 企业子客户all:自己和企业子客户| |参数的约束及描述：oneself：自身sub_customer: 企业子客户all:自己和企业子客户|
	Method *string `json:"method,omitempty"`
	// |参数名称：企业子账号ID。| |参数的约束及描述：注意：method不等于sub_customer的时候，该参数无效，如果method等于sub_customer，该参数不能为空|
	SubCustomerId *string `json:"sub_customer_id,omitempty"`
}

func (o QueryResRecordsDetailReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryResRecordsDetailReq struct{}"
	}

	return strings.Join([]string{"QueryResRecordsDetailReq", string(data)}, " ")
}
