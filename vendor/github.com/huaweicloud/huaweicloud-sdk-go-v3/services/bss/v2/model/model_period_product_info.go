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

type PeriodProductInfo struct {
	// |参数名称：ID标识| |参数约束及描述：同一次询价中不能重复，用于标识返回询价结果和请求的映射关系|
	Id string `json:"id"`
	// |参数名称：用户购买云服务产品的云服务类型| |参数约束及描述：例如EC2，云服务类型为hws.service.type.ec2|
	CloudServiceType string `json:"cloud_service_type"`
	// |参数名称：用户购买云服务产品的资源类型| |参数约束及描述：例如EC2中的VM，资源类型为hws.resource.type.vm。ResourceType是CloudServiceType中的一种资源，CloudServiceType由多种ResourceType组合提供|
	ResourceType string `json:"resource_type"`
	// |参数名称：用户购买云服务产品的资源规格| |参数约束及描述：例如VM的小型规格，资源规格为m1.tiny|
	ResourceSpec string `json:"resource_spec"`
	// |参数名称：云服务区编码| |参数约束及描述：云服务区编码|
	Region string `json:"region"`
	// |参数名称：可用区标识| |参数约束及描述：可用区标识|
	AvailableZone *string `json:"available_zone,omitempty"`
	// |参数名称：资源容量大小| |参数约束及描述：例如购买的卷大小或带宽大小，只有线性产品才有这个字段|
	ResourceSize *int32 `json:"resource_size,omitempty"`
	// |参数名称：资源容量度量标识| |参数约束及描述：枚举值如下：15：Mbps（购买带宽时使用）17：GB（购买云硬盘时使用）14：个只有线性产品才有这个字段|
	SizeMeasureId *int32 `json:"size_measure_id,omitempty"`
	// |参数名称：订购周期类型| |参数约束及描述：0：天；1：周；2：月；3：年；4：小时；|
	PeriodType int32 `json:"period_type"`
	// |参数名称：订购周期数| |参数约束及描述：订购周期数|
	PeriodNum int32 `json:"period_num"`
	// |参数名称：订购数量| |参数约束及描述：订购数量,有值时不能小于0|
	SubscriptionNum int32 `json:"subscription_num"`
}

func (o PeriodProductInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PeriodProductInfo struct{}"
	}

	return strings.Join([]string{"PeriodProductInfo", string(data)}, " ")
}
