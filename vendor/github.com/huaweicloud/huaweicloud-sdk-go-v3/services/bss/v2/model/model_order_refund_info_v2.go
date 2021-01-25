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

type OrderRefundInfoV2 struct {
	// |参数名称：该记录的ID。| |参数约束及描述：该记录的ID。|
	Id string `json:"id"`
	// |参数名称：金额。金额为负数，表示退订金额。金额为正数，表示已消费金额或收取的退订手续费。| |参数的约束及描述：金额。金额为负数，表示退订金额。金额为正数，表示已消费金额或收取的退订手续费。|
	Amount float32 `json:"amount"`
	// |参数名称：度量单位。1：元2：角3：分| |参数约束及描述：度量单位。1：元2：角3：分|
	MeasureId string `json:"measure_id"`
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。| |参数约束及描述：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。|
	ResourceTypeCode string `json:"resource_type_code"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	ServiceTypeCode string `json:"service_type_code"`
	// |参数名称：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。| |参数约束及描述：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。|
	RegionCode string `json:"region_code"`
	// |参数名称：退订金额、已消费金额或收取退订手续费对应的原订单ID。| |参数约束及描述：退订金额、已消费金额或收取退订手续费对应的原订单ID。|
	BaseOrderId *string `json:"base_order_id,omitempty"`
}

func (o OrderRefundInfoV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OrderRefundInfoV2 struct{}"
	}

	return strings.Join([]string{"OrderRefundInfoV2", string(data)}, " ")
}
