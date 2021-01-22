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

type CustomerOnDemandResource struct {
	// |参数名称：所属的客户ID。| |参数约束及描述：所属的客户ID。|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。| |参数约束及描述：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。|
	RegionCode *string `json:"region_code,omitempty"`
	// |参数名称：所属的AZ的编码。| |参数约束及描述：所属的AZ的编码。|
	AvailabilityZoneCode *string `json:"availability_zone_code,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。| |参数约束及描述：资源类型编码，例如ECS的VM为“hws.resource.type.vm”。具体请参见资源类型资源类型资源类型资源类型。|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：具体资源的ID。| |参数约束及描述：具体资源的ID。|
	ResourceId *string `json:"resource_id,omitempty"`
	// |参数名称：资源实例的名称。| |参数约束及描述：资源实例的名称。|
	ResourceName *string `json:"resource_name,omitempty"`
	// |参数名称：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：生效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	EffectiveTime *string `json:"effective_time,omitempty"`
	// |参数名称：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：失效时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：资源状态：1：正常（已开通）；2：宽限期；3：冻结中；4：变更中；5：正在关闭；6：已关闭。| |参数的约束及描述：资源状态：1：正常（已开通）；2：宽限期；3：冻结中；4：变更中；5：正在关闭；6：已关闭。|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：按需资源规格编码。| |参数约束及描述：按需资源规格编码。|
	ResourceSpecCode *string `json:"resource_spec_code,omitempty"`
	// |参数名称：资源容量大小。格式如| |参数约束及描述：资源容量大小。格式如：\"resourceInfo\": \"{\\\"specSize\\\":40.0}\"|
	ResourceInfo *string `json:"resource_info,omitempty"`
	// |参数名称：产品规格描述| |参数约束及描述：譬如虚拟机为：\"通用计算增强型|c6.2xlarge.4|8vCPUs|32GB|linux\"，硬盘为：\"云硬盘_SATA_LXH01|40.0GB\"|
	ProductSpecDesc *string `json:"product_spec_desc,omitempty"`
}

func (o CustomerOnDemandResource) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CustomerOnDemandResource struct{}"
	}

	return strings.Join([]string{"CustomerOnDemandResource", string(data)}, " ")
}
