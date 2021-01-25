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

type OrderInstanceV2 struct {
	// |参数名称：标识要开通资源的内部ID，资源开通以后生成的ID为resource_id。对应订购关系ID。| |参数约束及描述：标识要开通资源的内部ID，资源开通以后生成的ID为resource_id。对应订购关系ID。|
	Id *string `json:"id,omitempty"`
	// |参数名称：资源实例ID。| |参数约束及描述：资源实例ID。|
	ResourceId *string `json:"resource_id,omitempty"`
	// |参数名称：资源实例名。| |参数约束及描述：资源实例名。|
	ResourceName *string `json:"resource_name,omitempty"`
	// |参数名称：云服务资源池区域编码。| |参数约束及描述：云服务资源池区域编码。|
	RegionCode *string `json:"region_code,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。| |参数约束及描述：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：云服务产品的资源规格，例如VM的资源规格举例为“s2.small.1.linux”。具体请参见对应云服务的相关介绍。| |参数约束及描述：云服务产品的资源规格，例如VM的资源规格举例为“s2.small.1.linux”。具体请参见对应云服务的相关介绍。|
	ResourceSpecCode *string `json:"resource_spec_code,omitempty"`
	// |参数名称：资源项目ID。| |参数约束及描述：资源项目ID。|
	ProjectId *string `json:"project_id,omitempty"`
	// |参数名称：产品ID。| |参数约束及描述：产品ID。|
	ProductId *string `json:"product_id,omitempty"`
	// |参数名称：父资源实例ID。| |参数约束及描述：父资源实例ID。|
	ParentResourceId *string `json:"parent_resource_id,omitempty"`
	// |参数名称：是否是主资源。0：非主资源1：主资源| |参数的约束及描述：是否是主资源。0：非主资源1：主资源|
	IsMainResource *int32 `json:"is_main_resource,omitempty"`
	// |参数名称：资源状态：1：初始化2：已生效3：已过期4：已冻结5：宽限期6：冻结中7：冻结恢复中（预留，未启用）8：正在关闭| |参数的约束及描述：资源状态：1：初始化2：已生效3：已过期4：已冻结5：宽限期6：冻结中7：冻结恢复中（预留，未启用）8：正在关闭|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：资源生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。| |参数约束及描述：资源生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：资源过期时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。| |参数约束及描述：资源过期时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：到期策略：0：到期进入宽限期1：到期转按需2：到期后自动删除（从生效中直接删除）3：到期后自动续费4：到期后冻结5：到期后删除（从保留期删除）| |参数的约束及描述：到期策略：0：到期进入宽限期1：到期转按需2：到期后自动删除（从生效中直接删除）3：到期后自动续费4：到期后冻结5：到期后删除（从保留期删除）|
	ExpirePolicy *int32 `json:"expire_policy,omitempty"`
}

func (o OrderInstanceV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OrderInstanceV2 struct{}"
	}

	return strings.Join([]string{"OrderInstanceV2", string(data)}, " ")
}
