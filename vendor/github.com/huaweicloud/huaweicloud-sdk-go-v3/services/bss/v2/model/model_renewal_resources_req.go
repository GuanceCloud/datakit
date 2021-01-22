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

type RenewalResourcesReq struct {
	// |参数名称：资源ID列表。只支持传入主资源ID，最多100个资源ID。哪些资源是主资源请根据“2.1-查询客户包周期资源列表”接口响应参数中的“is_main_resource”来标识。| |参数约束以及描述：资源ID列表。只支持传入主资源ID，最多100个资源ID。哪些资源是主资源请根据“2.1-查询客户包周期资源列表”接口响应参数中的“is_main_resource”来标识。|
	ResourceIds []string `json:"resource_ids"`
	// |参数名称：周期类型：2：月；3：年| |参数的约束及描述：周期类型：2：月；3：年|
	PeriodType int32 `json:"period_type"`
	// |参数名称：周期数目：如果是月，目前支持1-11；如果是年，目前支持1-3| |参数的约束及描述：周期数目：如果是月，目前支持1-11；如果是年，目前支持1-3|
	PeriodNum int32 `json:"period_num"`
	// |参数名称：到期策略：0：进入宽限期1：转按需2：自动退订3：自动续订| |参数的约束及描述：到期策略：0：进入宽限期1：转按需2：自动退订3：自动续订|
	ExpirePolicy int32 `json:"expire_policy"`
	// |参数名称：是否自动支付。0：否1：是不填写的话，默认值是0，不自动支付。| |参数的约束及描述：是否自动支付。0：否1：是不填写的话，默认值是0，不自动支付。|
	IsAutoPay *int32 `json:"is_auto_pay,omitempty"`
}

func (o RenewalResourcesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RenewalResourcesReq struct{}"
	}

	return strings.Join([]string{"RenewalResourcesReq", string(data)}, " ")
}
