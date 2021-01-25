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

type ReclaimToPartnerAccountBalancesReq struct {
	// |参数名称：合作伙伴关联的客户的客户ID。| |参数约束及描述：合作伙伴关联的客户的客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：回收金额。| |参数的约束及描述：单位为元不能为负数，精确到小数点后两位。|
	Amount float32 `json:"amount"`
	// |参数名称：二级经销商ID。| |参数约束及描述：一级经销商回收二级经销商子客户余额时，需携带该字段。|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ReclaimToPartnerAccountBalancesReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimToPartnerAccountBalancesReq struct{}"
	}

	return strings.Join([]string{"ReclaimToPartnerAccountBalancesReq", string(data)}, " ")
}
