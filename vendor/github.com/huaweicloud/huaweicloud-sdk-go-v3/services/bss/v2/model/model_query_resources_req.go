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

type QueryResourcesReq struct {
	// |参数名称：资源ID列表。查询指定资源ID的资源（当only_main_resource=0时，查询指定资源及其附属资源）。最大支持50个ID同时作为条件查询。| |参数约束以及描述：资源ID列表。查询指定资源ID的资源（当only_main_resource=0时，查询指定资源及其附属资源）。最大支持50个ID同时作为条件查询。|
	ResourceIds *[]string `json:"resource_ids,omitempty"`
	// |参数名称：订单号。查询指定订单下的资源。| |参数约束及描述：订单号。查询指定订单下的资源。|
	OrderId *string `json:"order_id,omitempty"`
	// |参数名称：是否只查询主资源。0：查询主资源及附属资源。1：只查询主资源。默认值为0。| |参数的约束及描述：是否只查询主资源。0：查询主资源及附属资源。1：只查询主资源。默认值为0。|
	OnlyMainResource *int32 `json:"only_main_resource,omitempty"`
	// |参数名称：资源状态。查询指定状态的资源。1：初始化2：已生效3：已过期4：已冻结5：宽限期6：冻结中7：冻结恢复中（预留，未启用）8：正在关闭| |参数约束以及描述：资源状态。查询指定状态的资源。1：初始化2：已生效3：已过期4：已冻结5：宽限期6：冻结中7：冻结恢复中（预留，未启用）8：正在关闭|
	StatusList *[]int32 `json:"status_list,omitempty"`
	// |参数名称：偏移量，从0开始默认值是0。| |参数的约束及描述：偏移量，从0开始默认值是0。|
	Offset *int32 `json:"offset,omitempty"`
	// |参数名称：每次查询的条数。默认值是10。最大值是500。| |参数的约束及描述：每次查询的条数。默认值是10。最大值是500。|
	Limit *int32 `json:"limit,omitempty"`
}

func (o QueryResourcesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryResourcesReq struct{}"
	}

	return strings.Join([]string{"QueryResourcesReq", string(data)}, " ")
}
