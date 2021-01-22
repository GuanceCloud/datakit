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

type QueryCustomersBalancesReq struct {
	// |参数名称：客户的客户信息列表。| |参数约束以及描述：客户的客户信息列表。|
	CustomerInfos []CustomerInfoV2 `json:"customer_infos"`
	// |参数名称：二级经销商ID。| |参数约束及描述：查询二级经销商子客户的账户余额的时候，需要携带这个字段。|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o QueryCustomersBalancesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryCustomersBalancesReq struct{}"
	}

	return strings.Join([]string{"QueryCustomersBalancesReq", string(data)}, " ")
}
