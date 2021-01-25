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

type QueryCustomerOnDemandResourcesReq struct {
	// |参数名称：所属的客户ID。| |参数约束及描述：所属的客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。| |参数约束及描述：云服务区编码，例如：“cn-north-1”。具体请参见地区和终端节点地区和终端节点对应云服务的“区域”列的值。|
	RegionCode *string `json:"region_code,omitempty"`
	// |参数名称：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。| |参数约束及描述：云服务类型编码，例如ECS的云服务类型编码为“hws.service.type.ec2”。具体请参见云服务类型云服务类型云服务类型云服务类型。|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：资源ID批量查询| |参数约束以及描述：用于查询指定资源ID对应的资源。最多支持同时传递50个Id的列表。|
	ResourceIds *[]string `json:"resource_ids,omitempty"`
	// |参数名称：生效时间的开始时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：生效时间的开始时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	EffectiveTimeBegin *string `json:"effective_time_begin,omitempty"`
	// |参数名称：生效时间的结束时间UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：生效时间的结束时间UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。其中，HH范围是0～23，mm和ss范围是0～59。|
	EffectiveTimeEnd *string `json:"effective_time_end,omitempty"`
	// |参数名称：偏移量，从0开始。默认值：0| |参数的约束及描述：偏移量，从0开始。默认值：0|
	Offset *int32 `json:"offset,omitempty"`
	// |参数名称：一次查询的条数，默认10条。| |参数的约束及描述：一次查询的条数，默认10条。|
	Limit *int32 `json:"limit,omitempty"`
	// |参数名称：资源状态：1：正常（已开通）；2：宽限期；3：冻结中；4：变更中；5：正在关闭；6：已关闭。| |参数的约束及描述：资源状态：1：正常（已开通）；2：宽限期；3：冻结中；4：变更中；5：正在关闭；6：已关闭。|
	Status *int32 `json:"status,omitempty"`
	// |参数名称：二级经销商ID，如果想查询二级经销商的子客户的资源列表，必须携带该字段，否则只能查询自己的子客户的按需资源| |参数约束及描述：二级经销商ID，如果想查询二级经销商的子客户的资源列表，必须携带该字段，否则只能查询自己的子客户的按需资源|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o QueryCustomerOnDemandResourcesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryCustomerOnDemandResourcesReq struct{}"
	}

	return strings.Join([]string{"QueryCustomerOnDemandResourcesReq", string(data)}, " ")
}
