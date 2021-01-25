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

type RateOnPeriodReq struct {
	// |参数名称：项目ID| |参数约束及描述：如果使用客户AK/SK或者Token，可以调用“通过assume_role方式获取用户token”接口获取“regionId”取值对应的project id。|
	ProjectId string `json:"project_id"`
	// |参数名称：产品信息列表| |参数的约束及描述：询价时要询价产品的信息的列表|
	ProductInfos []PeriodProductInfo `json:"product_infos"`
}

func (o RateOnPeriodReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RateOnPeriodReq struct{}"
	}

	return strings.Join([]string{"RateOnPeriodReq", string(data)}, " ")
}
