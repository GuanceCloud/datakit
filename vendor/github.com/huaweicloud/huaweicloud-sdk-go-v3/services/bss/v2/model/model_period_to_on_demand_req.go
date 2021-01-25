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

type PeriodToOnDemandReq struct {
	// |参数名称：动作| |参数约束及描述：动作 动作SET_UP：设置CANCEL：取消|
	Operation string `json:"operation"`
	// |参数名称：资源ID| |参数约束以及描述：资源ID 资源ID。您可以调用“2.1-查询客户包年/包月资源列表”接口获取资源ID。只支持传入主资源ID，最多100个资源ID。设置的时候，主资源和对应的子资源一起转按需。哪些资源是主资源请根据“2.1-查询客户包年/包月资源列表”接口响应参数中的“is_main_resource”来标识。|
	ResourceIds []string `json:"resource_ids"`
}

func (o PeriodToOnDemandReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PeriodToOnDemandReq struct{}"
	}

	return strings.Join([]string{"PeriodToOnDemandReq", string(data)}, " ")
}
